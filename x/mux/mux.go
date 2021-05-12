// Package mux provides a simple Discord message route multiplexer that
// parses messages and then executes a matching registered handler, if found.
// mux can be used with both Disgord and the DiscordGo library.
package mux

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/utils"
)

// Route holds information about a specific message route handler
type Route struct {
	Pattern     string      // match pattern that should trigger this route handler
	Description string      // short description of this route
	Help        string      // detailed help string for this route
	Run         HandlerFunc // route handler function to call
	Access      int         // access level for the command
}

// Context holds a bit of extra data we pass along to route handlers
// This way processing some of this only needs to happen once.
type Context struct {
	Fields          []string
	Content         string
	IsDirected      bool
	IsPrivate       bool
	HasPrefix       bool
	HasMention      bool
	HasMentionFirst bool
}

// HandlerFunc is the function signature required for a message route handler.
type HandlerFunc func(*discordgo.Session, *discordgo.Message, *Context)

// Mux is the main struct for all mux methods.
type Mux struct {
	Routes  []*Route
	Default *Route
	Prefix  string
}

// New returns a new Discord message route mux
func New() *Mux {
	m := &Mux{}
	m.Prefix = "-db "
	return m
}

// Route allows you to register a route
func (m *Mux) Route(pattern, desc string, cb HandlerFunc, access int) (*Route, error) {

	r := Route{}
	r.Pattern = pattern
	r.Description = desc
	r.Run = cb
	r.Access = access
	m.Routes = append(m.Routes, &r)

	fmt.Printf("Command \"%s%s\" loaded\n", m.Prefix, r.Pattern)

	return &r, nil
}

// FuzzyMatch attempts to find the best route match for a given message.
func (m *Mux) FuzzyMatch(msg string) (*Route, []string) {

	// Tokenize the msg string into a slice of words
	fields := strings.Fields(msg)

	// no point to continue if there's no fields
	if len(fields) == 0 {
		return nil, nil
	}

	// Search though the command list for a match
	var r *Route
	// var rank int

	var fk int
	for fk, fv := range fields {

		for _, rv := range m.Routes {

			// If we find an exact match, return that immediately.
			if rv.Pattern == fv {
				return rv, fields[fk:]
			}

			// Some "Fuzzy" searching...
			// if strings.HasPrefix(rv.Pattern, fv) {
			// 	if len(fv) > rank {
			// 		r = rv
			// 		rank = len(fv)
			// 	}
			// }
		}
	}
	return r, fields[fk:]
}

type MessageLog struct {
	ChannelName     string
	UserName        string
	Content         string
	OriginalContent string                         `bson:"originalContent,omitempty"`
	Attachments     []*discordgo.MessageAttachment `bson:"attachments,omitempty"`
	MessageID       string
	Action          string
	Time            time.Time
}

func ImageCopyEmbeds(ds *discordgo.Session, msg *discordgo.Message) []*discordgo.MessageEmbed {
	channel, _ := ds.Channel(msg.ChannelID)

	embeds := []*discordgo.MessageEmbed{}

	for _, attachment := range msg.Attachments {
		embed := &discordgo.MessageEmbed{
			Color: 3066993,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: msg.Author.AvatarURL("64"),
			},
			Author: &discordgo.MessageEmbedAuthor{
				Name: msg.Author.Username,
			},
			Description: "File: " + attachment.URL + "\n\nMessage: " + config.MessageLink(msg),
			Footer: &discordgo.MessageEmbedFooter{
				Text: channel.Name,
			},
			Timestamp: string(msg.Timestamp),
		}

		embed.Image = &discordgo.MessageEmbedImage{
			URL:      attachment.URL,
			ProxyURL: attachment.ProxyURL,
		}
		embed.Video = &discordgo.MessageEmbedVideo{
			URL: attachment.URL,
		}
		embeds = append(embeds, embed)
	}

	return embeds
}

// OnMessageCreate is a DiscordGo Event Handler function.  This must be
// registered using the DiscordGo.Session.AddHandler function.  This function
// will receive all Discord messages and parse them for matches to registered
// routes.
func (m *Mux) OnMessageCreate(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	session := mongo.MDB.Clone()
	defer session.Close()
	db := session.DB(mongo.DB_NAME)

	abort := m.EnsureSticky(db, ds, mc.Message)
	if abort {
		return
	}

	// Copy images so they can still be referenced
	if mc.Author.ID != ds.State.User.ID && mc.GuildID == "755437328515989564" && len(mc.Message.Attachments) > 0 {
		utils.PrintJSON(mc.Message)
		embeds := ImageCopyEmbeds(ds, mc.Message)
		for _, embed := range embeds {
			imageDumpChannel := "840804942326530088"
			ds.ChannelMessageSendEmbed(imageDumpChannel, embed)
			// ds.ChannelMessageSendComplex(imageDumpChannel, &discordgo.MessageSend{
			// 	Content: mc.Message.Attachments[0].URL,
			// 	Embed:   embed,
			// })
		}
	}

	m.LogMessageCreate(db, ds, mc, nil)

	// fmt.Println("Got message", mc.Message.MessageReference, mc.Message.MessageReference != nil)
	if mc.Message.MessageReference != nil {
		// Get the highest message up the reply chain, and check if it was the bot
		wasBotReply := false
		ref := mc.Message.MessageReference
		for {
			msg, err := ds.ChannelMessage(ref.ChannelID, ref.MessageID)
			if err != nil {
				log.Printf("Failed to get reply message: %s", err)
				break
			}
			if msg.MessageReference == nil {
				wasBotReply = msg.Author.ID == ds.State.User.ID
				break
			}
			ref = msg.MessageReference
		}

		if wasBotReply {
			tweetConfig := config.MaybeGetTweetConfig(ref.ChannelID)
			if tweetConfig != nil {
				m.DoTweetUpdateByReply(ds, mc.Message, ref)
				return
			}
		}
	}

	var err error

	// Ignore all messages created by the Bot account itself
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	// Handle Youtube copy pipelines
	for _, cp := range config.CopyPipelines {
		if cp.ChannelID == mc.ChannelID {
			go m.CopyMessageToYoutube(ds, mc.Message, cp)
		}
	}

	// if mc.Content == config.Emoji("delucringe") {
	// 	doDelete := false
	// 	if mc.Message.MessageReference != nil {
	// 		rMsg, err := ds.ChannelMessage(mc.Message.MessageReference.ChannelID, mc.Message.MessageReference.MessageID)
	// 		if err != nil {
	// 			log.Printf("Failed to get reply message: %s", err)
	// 		} else if rMsg.Content == config.Emoji("delucringe") && rMsg.Author.ID == ds.State.User.ID {
	// 			doDelete = true
	// 		}
	// 	}

	// 	if doDelete {
	// 		ds.ChannelMessageDelete(mc.Message.ChannelID, mc.Message.ID)
	// 	} else {
	// 		ds.ChannelMessageSendReply(mc.Message.ChannelID, config.Emoji("delucringe"), mc.Message.Reference())
	// 	}
	// }

	channelName := ""
	channel, err := ds.Channel(mc.ChannelID)
	if err == nil {
		channelName = channel.Name
	} else {
		channelName = err.Error()
	}

	// 1 in 20000 chance for a :delucringe: response on every message
	randNum := rand.Intn(20000)
	// fmt.Println("Check random", randNum)
	if randNum == 0 && channel.ParentID != "779849308525690900" {
		msg, _ := ds.ChannelMessageSendReply(mc.Message.ChannelID, config.Emoji("delucringe"), mc.Message.Reference())
		// ds.ChannelMessageSend("", fmt.Sprintf("Cringed! %s", msg.))
		fmt.Println("Cringed!", msg)
	}

	// Ignore all messages by non-moderators
	msg := mc.Author.Username + ": " + mc.Content
	if IsModerator(ds, mc) {
		msg = "[M] " + msg
	}
	fmt.Println("#" + channelName + " - " + msg)

	// Create Context struct that we can put various infos into
	ctx := &Context{
		Content: strings.TrimSpace(mc.Content),
	}

	// Catch the special modmail "=v" command, alias it to a real command
	if strings.HasPrefix(ctx.Content, "=v") {
		if config.IsModmailChannel(ds, mc.GuildID, mc.ChannelID) {
			ctx.Content = strings.TrimSpace(m.Prefix) + " v" + strings.TrimPrefix(ctx.Content, "=v")
		}
	}

	// Catch the special modmail "=vd" command, alias it to a real command
	if strings.HasPrefix(ctx.Content, "=vd") {
		if config.IsModmailChannel(ds, mc.GuildID, mc.ChannelID) {
			ctx.Content = strings.TrimSpace(m.Prefix) + " vd" + strings.TrimPrefix(ctx.Content, "=vd")
		}
	}

	// Catch the special "!clear" command, alias it to a real command
	if strings.HasPrefix(ctx.Content, "!clear") {
		ctx.Content = strings.TrimSpace(m.Prefix) + " clear" + strings.TrimPrefix(ctx.Content, "!clear")
	}

	// Fetch the channel for this Message
	var c *discordgo.Channel
	c, err = ds.State.Channel(mc.ChannelID)
	if err != nil {
		// Try fetching via REST API
		c, err = ds.Channel(mc.ChannelID)
		if err != nil {
			log.Printf("unable to fetch Channel for Message, %s", err)
		} else {
			// Attempt to add this channel into our State
			err = ds.State.ChannelAdd(c)
			if err != nil {
				log.Printf("error updating State with Channel, %s", err)
			}
		}
	}
	// Add Channel info into Context (if we successfully got the channel)
	if c != nil {
		if c.Type == discordgo.ChannelTypeDM {
			ctx.IsPrivate, ctx.IsDirected = true, true
		}
	}

	// Detect @name or @nick mentions
	// if !ctx.IsDirected && mc.Mentions != nil {

	// 	// Detect if Bot was @mentioned
	// 	for _, v := range mc.Mentions {

	// 		if v.ID == ds.State.User.ID {

	// 			ctx.IsDirected, ctx.HasMention = true, true

	// 			reg := regexp.MustCompile(fmt.Sprintf("<@!?(%s)>", ds.State.User.ID))

	// 			// Was the @mention the first part of the string?
	// 			if reg.FindStringIndex(ctx.Content)[0] == 0 {
	// 				ctx.HasMentionFirst = true
	// 			}

	// 			// strip bot mention tags from content string
	// 			ctx.Content = reg.ReplaceAllString(ctx.Content, "")

	// 			break
	// 		}
	// 	}
	// }

	// Detect prefix mention
	if !ctx.IsDirected && len(m.Prefix) > 0 {

		// TODO : Must be changed to support a per-guild user defined prefix
		if strings.HasPrefix(ctx.Content, m.Prefix) {
			ctx.IsDirected, ctx.HasPrefix, ctx.HasMentionFirst = true, true, true
			ctx.Content = strings.TrimPrefix(ctx.Content, m.Prefix)
		}
	}

	// For now, if we're not specifically mentioned we do nothing.
	// later I might add an option for global non-mentioned command words
	if !ctx.IsDirected {
		return
	}

	// Try to find the "best match" command out of the message.
	fmt.Println("Received command:", ctx.Content)
	r, fl := m.FuzzyMatch(ctx.Content)
	if r != nil {
		if !HasAccess(ds, mc, r.Access) {
			return
		}

		ctx.Fields = fl
		r.Run(ds, mc.Message, ctx)
		return
	}

	// If no command match was found, call the default.
	// Ignore if only @mentioned in the middle of a message
	if m.Default != nil && (ctx.HasMentionFirst) {
		// TODO: This could use a ratelimit
		// or should the ratelimit be inside the cmd handler?..
		// In the case of "talking" to another bot, this can create an endless
		// loop.  Probably most common in private messages.
		m.Default.Run(ds, mc.Message, ctx)
	}

}

func (m *Mux) OnMessageDelete(ds *discordgo.Session, md *discordgo.MessageDelete) {
	session := mongo.MDB.Clone()
	defer session.Close()
	db := session.DB(mongo.DB_NAME)
	mlog := db.C("message_logs")

	channelName := ""
	channel, err := ds.Channel(md.ChannelID)
	if err == nil {
		channelName = channel.Name
	} else {
		channelName = err.Error()
	}

	deleted := MessageLog{}
	err = mlog.Find(bson.M{"messageid": md.ID}).One(&deleted)
	if err != nil {
		fmt.Println("Failed to get deleted message: " + err.Error())
	}
	log.Printf("DELETED MESSAGE - %s: %s", deleted.UserName, deleted.Content)

	mlog.Insert(MessageLog{
		ChannelName: channelName,
		UserName:    deleted.UserName,
		Content:     deleted.Content,
		MessageID:   md.ID,
		Action:      "delete",
		Time:        time.Now(),
	})
}

func (m *Mux) OnMessageDeleteBulk(ds *discordgo.Session, mdb *discordgo.MessageDeleteBulk) {
	session := mongo.MDB.Clone()
	defer session.Close()
	db := session.DB(mongo.DB_NAME)
	mlog := db.C("message_logs")

	channelName := ""
	channel, err := ds.Channel(mdb.ChannelID)
	if err == nil {
		channelName = channel.Name
	} else {
		channelName = err.Error()
	}

	for _, msgID := range mdb.Messages {
		deleted := MessageLog{}
		err = mlog.Find(bson.M{"messageid": msgID}).One(&deleted)
		if err != nil {
			fmt.Println("Failed to get bulk deleted message: " + err.Error())
		}
		log.Printf("BULK DELETED MESSAGE - %s: %s", deleted.UserName, deleted.Content)

		mlog.Insert(MessageLog{
			ChannelName: channelName,
			UserName:    deleted.UserName,
			Content:     deleted.Content,
			MessageID:   msgID,
			Action:      "delete",
			Time:        time.Now(),
		})
	}
}

func (m *Mux) OnMessageUpdate(ds *discordgo.Session, mu *discordgo.MessageUpdate) {
	session := mongo.MDB.Clone()
	defer session.Close()
	db := session.DB(mongo.DB_NAME)
	mlog := db.C("message_logs")

	channelName := ""
	channel, err := ds.Channel(mu.ChannelID)
	if err == nil {
		channelName = channel.Name
	} else {
		channelName = err.Error()
	}

	mlog.Insert(MessageLog{
		ChannelName: channelName,
		UserName:    "#",
		Content:     mu.Content,
		MessageID:   mu.ID,
		Action:      "edit",
		Time:        time.Now(),
	})
}

func (m *Mux) LogMessageCreate(db *mgo.Database, ds *discordgo.Session, mc *discordgo.MessageCreate, channelName *string) {
	mlog := db.C("message_logs")
	if channelName == nil {
		name := ""
		channel, err := ds.Channel(mc.ChannelID)
		if err == nil {
			name = channel.Name
		} else {
			name = err.Error()
		}
		channelName = &name
	}

	mlog.Insert(MessageLog{
		ChannelName: *channelName,
		UserName:    mc.Author.Username + "#" + mc.Author.Discriminator,
		Content:     mc.Content,
		Attachments: mc.Attachments,
		MessageID:   mc.ID,
		Action:      "create",
		Time:        time.Now(),
	})
}

var Pushpin = "\U0001F4CC"

func (m *Mux) AddReaction(ds *discordgo.Session, ra *discordgo.MessageReactionAdd) {
	// Don't react to the bot's own reactions
	if ra.UserID == ds.State.User.ID {
		return
	}
	if channelID, ok := config.Proposals[ra.MessageID]; ok {
		m.UpdateProposal(ds, ra.GuildID, channelID, ra.MessageID)
	}

	if tu, ok := config.TweetUpdates[ra.MessageID]; ok {
		if ra.Emoji.Name == "✅" {
			m.DoTweetUpdate(ds, tu)
		}
		if ra.Emoji.Name == "❌" {
			m.CancelTweetUpdate(ds, tu)
		}
	}

	if e, ok := config.Extractions[ra.MessageID]; ok {
		if ra.Emoji.Name == "🇾" {
			m.DoExtraction(ds, e)
		}
		if ra.Emoji.Name == "🇳" {
			m.CancelExtraction(ds, e)
		}
	}

	if ra.Emoji.Name == "📌" && IsStaff(ds, ra.GuildID, ra.UserID) {
		err := ds.ChannelMessagePin(ra.ChannelID, ra.MessageID)
		if err != nil {
			ds.ChannelMessageSend(ra.ChannelID, fmt.Sprintf("Failed to pin message: %s", err))
		}
	}
}

func (m *Mux) RemoveReaction(ds *discordgo.Session, rr *discordgo.MessageReactionRemove) {
	// Don't react to the bot's own reactions
	if rr.UserID == ds.State.User.ID {
		return
	}
	if channelID, ok := config.Proposals[rr.MessageID]; ok {
		m.UpdateProposal(ds, rr.GuildID, channelID, rr.MessageID)
	}

	if rr.Emoji.Name == "📌" && IsStaff(ds, rr.GuildID, rr.UserID) {
		users, err := ds.MessageReactions(rr.ChannelID, rr.MessageID, Pushpin, 100, "", "")
		if err != nil {
			ds.ChannelMessageSend(rr.ChannelID, fmt.Sprintf("Failed to get reactions: %s", err))
			return
		}
		for _, user := range users {
			if IsStaff(ds, rr.GuildID, user.ID) {
				// There's still a staff pin so don't unpin
				return
			}
		}

		err = ds.ChannelMessageUnpin(rr.ChannelID, rr.MessageID)
		if err != nil {
			ds.ChannelMessageSend(rr.ChannelID, fmt.Sprintf("Failed to unpin message: %s", err))
		}
	}
}

// Package mux provides a simple Discord message route multiplexer that
// parses messages and then executes a matching registered handler, if found.
// mux can be used with both Disgord and the DiscordGo library.
package mux

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
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
func (m *Mux) Route(pattern, desc string, cb HandlerFunc) (*Route, error) {

	r := Route{}
	r.Pattern = pattern
	r.Description = desc
	r.Run = cb
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
	var rank int

	var fk int
	for fk, fv := range fields {

		for _, rv := range m.Routes {

			// If we find an exact match, return that immediately.
			if rv.Pattern == fv {
				return rv, fields[fk:]
			}

			// Some "Fuzzy" searching...
			if strings.HasPrefix(rv.Pattern, fv) {
				if len(fv) > rank {
					r = rv
					rank = len(fv)
				}
			}
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

// OnMessageCreate is a DiscordGo Event Handler function.  This must be
// registered using the DiscordGo.Session.AddHandler function.  This function
// will receive all Discord messages and parse them for matches to registered
// routes.
func (m *Mux) OnMessageCreate(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	session := mongo.MDB.Clone()
	defer session.Close()
	db := session.DB(mongo.DB_NAME)

	m.LogMessageCreate(db, ds, mc, nil)

	var err error

	// Ignore all messages created by the Bot account itself
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	channelName := ""
	channel, err := ds.Channel(mc.ChannelID)
	if err == nil {
		channelName = channel.Name
	} else {
		channelName = err.Error()
	}

	// Ignore all messages by non-moderators
	msg := mc.Author.Username + ": " + mc.Content
	if IsModerator(ds, mc) {
		msg = "[M] " + msg
	}
	fmt.Println("#" + channelName + " - " + msg)
	if !IsModerator(ds, mc) {
		return
	}

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
	if !ctx.IsDirected {

		// Detect if Bot was @mentioned
		for _, v := range mc.Mentions {

			if v.ID == ds.State.User.ID {

				ctx.IsDirected, ctx.HasMention = true, true

				reg := regexp.MustCompile(fmt.Sprintf("<@!?(%s)>", ds.State.User.ID))

				// Was the @mention the first part of the string?
				if reg.FindStringIndex(ctx.Content)[0] == 0 {
					ctx.HasMentionFirst = true
				}

				// strip bot mention tags from content string
				ctx.Content = reg.ReplaceAllString(ctx.Content, "")

				break
			}
		}
	}

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

	message, err := ds.ChannelMessage(md.ChannelID, md.ID)
	if err != nil {
		fmt.Println("Failed to get bulk deleted message: " + err.Error())
	}
	utils.PrintJSON(message)

	mlog.Insert(MessageLog{
		ChannelName: channelName,
		UserName:    "#",
		Content:     md.Content,
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
		message, err := ds.ChannelMessage(mdb.ChannelID, msgID)
		if err != nil {
			fmt.Println("Failed to get bulk deleted message: " + err.Error())
		}
		mlog.Insert(MessageLog{
			ChannelName: channelName,
			UserName:    "#",
			Content:     message.Content,
			MessageID:   message.ID,
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
}

func (m *Mux) RemoveReaction(ds *discordgo.Session, rr *discordgo.MessageReactionRemove) {
	// Don't react to the bot's own reactions
	if rr.UserID == ds.State.User.ID {
		return
	}
	if channelID, ok := config.Proposals[rr.MessageID]; ok {
		m.UpdateProposal(ds, rr.GuildID, channelID, rr.MessageID)
	}
}

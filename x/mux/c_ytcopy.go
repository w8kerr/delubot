package mux

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/youtubesvc"
)

var YTSvc *youtubesvc.UserYoutubeService

func (m *Mux) YoutubeCopy(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	ctx.Content = strings.TrimPrefix(ctx.Content, "ytcopy")
	ctx.Content = strings.TrimSpace(ctx.Content)

	if ctx.Content == "" {
		respond("🔺Usage: -db ytcopy <youtube video ID or link>")
		return
	}

	parts := strings.Split(ctx.Content, " ")
	link := parts[0]
	prefix := strings.Join(parts[1:], " ")

	if prefix == "" {
		prefix = "[Discord🤖|EN] "
	}

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	c := context.Background()
	c = context.WithValue(c, "mgo", session)

	svc, err := youtubesvc.NewYoutubeService(c)

	videoID, err := svc.ParseVideoID(link)
	if err != nil {
		respond("🔺That doesn't like a Youtube link or video ID to me!")
		return
	}

	livechatID, videoTitle, err := svc.GetLivechatID(videoID)
	if err != nil {
		respond(fmt.Sprintf("🔺Failed to connect to live chat: %s", err))
		return
	}

	// Save the copy pipeline
	cp := config.CopyPipeline{
		CreatedAt:         time.Now(),
		CreatedBy:         dm.Author.ID,
		CreatedByName:     dm.Author.Username,
		Type:              "youtube",
		ChannelID:         dm.ChannelID,
		Prefix:            prefix,
		YoutubeVideoID:    videoID,
		YoutubeVideoTitle: videoTitle,
		YoutubeLivechatID: livechatID,
	}
	err = config.SetCopyPipeline(cp)
	if err != nil {
		respond(fmt.Sprintf("🔺Failed to initialize copy pipeline: %s", err))
		return
	}

	if YTSvc == nil {
		YTSvc, err = youtubesvc.NewUserYoutubeService(config.YoutubeOauthToken, &config.YoutubeRefreshToken)
		if err != nil {
			fmt.Println("Error", err)
		}
	}

	ds.ChannelMessageSendEmbed(dm.ChannelID, StartCopyEmbed(cp))
}

// func (m *Mux) EnsureCopy(db *mgo.Database, ds *discordgo.Session, dm *discordgo.Message) bool {

// }

func (m *Mux) EndYoutubeCopy(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	_, err := config.RemoveCopyPipeline(dm.ChannelID)
	if err != nil {
		respond(fmt.Sprintf("🔺Failed to initialize copy pipeline: %s", err))
		return
	}

	err = ds.MessageReactionAdd(dm.ChannelID, dm.ID, "\U0001F44D")
	if err != nil {
		fmt.Printf("🔺Failed to add \U0001F44D reaction, %s\n", err)
	}
}

func StartCopyEmbed(cp config.CopyPipeline) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color: 3066993,
		// Thumbnail: &discordgo.MessageEmbedThumbnail{
		// 	URL: sticky.AuthorAvatarURL,
		// },
		// Author: &discordgo.MessageEmbedAuthor{
		// 	Name: sticky.AuthorName,
		// },
		Description: fmt.Sprintf("Now copying messages in this channel to \"\"\nPrefix: `%s`\nType `-db endcopy` to end", cp.Prefix),
		Footer: &discordgo.MessageEmbedFooter{
			Text: cp.CreatedByName,
		},
		Timestamp: cp.CreatedAt.Format(time.RFC3339),
	}
	return embed
}

func (m *Mux) CopyMessageToYoutube(ds *discordgo.Session, dm *discordgo.Message, cp config.CopyPipeline) {
	respond := GetResponder(ds, dm)

	var err error
	if YTSvc == nil {
		YTSvc, err = youtubesvc.NewUserYoutubeService(config.YoutubeOauthToken, &config.YoutubeRefreshToken)
		if err != nil {
			fmt.Println("Error", err)
		}
	}

	text, err := dm.ContentWithMoreMentionsReplaced(ds)
	if err != nil {
		respond(fmt.Sprintf("🔺Failed to copy message to Youtube: %s", err))
		return
	}

	text, abort := RemoveUnwantedElements(text)
	if abort {
		return
	}

	characterLimit := 200 - len(cp.Prefix)

	words := strings.Split(text, " ")

	output := ""
	for i := 0; i < len(words); i++ {
		nextWord := words[i]
		if output != "" {
			nextWord = " " + nextWord
		}
		if len(output)+len(nextWord) > characterLimit {
			_, err = YTSvc.SendChatMessage(cp.YoutubeLivechatID, cp.Prefix+output)
			fmt.Println("Sent youtube message,", cp.Prefix+output)
			if err != nil {
				respond(fmt.Sprintf("🔺Failed to copy message to Youtube: %s", err))
				return
			}
			time.Sleep(1 * time.Second)
			output = nextWord
		} else {
			output += nextWord
		}
	}
	_, err = YTSvc.SendChatMessage(cp.YoutubeLivechatID, cp.Prefix+output)
	fmt.Println("Sent youtube message,", cp.Prefix+output)
	if err != nil {
		respond(fmt.Sprintf("🔺Failed to copy message to Youtube: %s", err))
		return
	}
}

func RemoveUnwantedElements(text string) (string, bool) {
	if strings.HasPrefix(text, "-db") {
		return "", true
	}
	return text, false
}

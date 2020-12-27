package mux

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/models"
	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/tweetsync"
)

func (m *Mux) TweetTranslate(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	foundChannel := false
	for _, tsc := range config.TweetSyncChannels {
		if tsc.ChannelID == dm.ChannelID {
			foundChannel = true
		}
	}

	if !foundChannel {
		respond("🔺This command can only be used in a Twitter sync channel")
		return
	}

	ctx.Content = strings.TrimPrefix(ctx.Content, "ttl")
	ctx.Content = strings.TrimSpace(ctx.Content)

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)
	stCol := db.C("synced_tweets")

	st := models.SyncedTweet{}
	err := stCol.Find(bson.M{
		"human_translated": false,
		"created_at":       bson.M{"$gte": time.Now().Add(-24 * time.Hour)},
	}).Sort("created_at").Limit(1).One(&st)
	if err != nil {
		msg := respond(fmt.Sprintf("🔺Failed to get earliest untranslated Tweet, %s", err))
		err = ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\u274C")
		if err != nil {
			fmt.Printf("Failed to add \u274C reaction, %s\n", err)
		}
		config.TweetUpdates[msg.ID] = config.TweetUpdate{
			ChannelID:     msg.ChannelID,
			UserMessageID: dm.ID,
			BotMessageID:  msg.ID,
		}
		return
	}

	if ctx.Content == "" {
		msg := respond(fmt.Sprintf("🔺Usage: -db ttl <translation for oldest untranslated Tweet within 24 hours>\nCurrently pointing to:\n❝ %s ❞", st.Tweet.Text))
		err = ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\u274C")
		if err != nil {
			fmt.Printf("Failed to add \u274C reaction, %s\n", err)
		}
		config.TweetUpdates[msg.ID] = config.TweetUpdate{
			ChannelID:     msg.ChannelID,
			UserMessageID: dm.ID,
			BotMessageID:  msg.ID,
		}
		return
	}

	msg := respond(fmt.Sprintf("🔺Translate:\n❝ %s ❞\nto\n❝ %s ❞", st.Tweet.Text, ctx.Content))
	err = ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\u2705")
	if err != nil {
		fmt.Printf("Failed to add \u2705 reaction, %s\n", err)
	}
	err = ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\u274C")
	if err != nil {
		fmt.Printf("Failed to add \u274C reaction, %s\n", err)
	}
	config.TweetUpdates[msg.ID] = config.TweetUpdate{
		ChannelID:      msg.ChannelID,
		UserMessageID:  dm.ID,
		BotMessageID:   msg.ID,
		TweetMessageID: st.MessageID,
		Translation:    ctx.Content,
		Translator:     dm.Author.Username,
	}
}

func (m *Mux) TweetEdit(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	foundChannel := false
	for _, tsc := range config.TweetSyncChannels {
		if tsc.ChannelID == dm.ChannelID {
			foundChannel = true
		}
	}

	if !foundChannel {
		respond("🔺This command can only be used in a Twitter sync channel")
		return
	}

	ctx.Content = strings.TrimPrefix(ctx.Content, "tedit")
	ctx.Content = strings.TrimSpace(ctx.Content)

	if ctx.Content == "" {
		msg := respond("🔺Usage: -db tedit <number of tweet counting upwards> <translation>")
		err := ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\u274C")
		if err != nil {
			fmt.Printf("Failed to add \u274C reaction, %s\n", err)
		}
		config.TweetUpdates[msg.ID] = config.TweetUpdate{
			ChannelID:     msg.ChannelID,
			UserMessageID: dm.ID,
			BotMessageID:  msg.ID,
		}
		return
	}

	parts := strings.Split(ctx.Content, " ")
	numStr := parts[0]
	num, err := strconv.Atoi(numStr)
	if err != nil || num == 0 {
		msg := respond("🔺Usage: -db tedit <number of tweet counting upwards> <translation>")
		err := ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\u274C")
		if err != nil {
			fmt.Printf("Failed to add \u274C reaction, %s\n", err)
		}
		config.TweetUpdates[msg.ID] = config.TweetUpdate{
			ChannelID:     msg.ChannelID,
			UserMessageID: dm.ID,
			BotMessageID:  msg.ID,
		}
		return
	}

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)
	stCol := db.C("synced_tweets")

	st := models.SyncedTweet{}
	err = stCol.Find(bson.M{}).Sort("-created_at").Skip(num - 1).Limit(1).One(&st)
	if err != nil {
		msg := respond(fmt.Sprintf("🔺Failed to get earliest untranslated Tweet, %s", err))
		err = ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\u274C")
		if err != nil {
			fmt.Printf("Failed to add \u274C reaction, %s\n", err)
		}
		config.TweetUpdates[msg.ID] = config.TweetUpdate{
			ChannelID:     msg.ChannelID,
			UserMessageID: dm.ID,
			BotMessageID:  msg.ID,
		}
		return
	}

	if len(parts) == 1 {
		msg := respond(fmt.Sprintf("🔺Usage: -db tedit <number of tweet counting upwards> <translation>\nCurrently pointing to:\n❝ %s ❞", st.Tweet.Text))
		err = ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\u274C")
		if err != nil {
			fmt.Printf("Failed to add \u274C reaction, %s\n", err)
		}
		config.TweetUpdates[msg.ID] = config.TweetUpdate{
			ChannelID:     msg.ChannelID,
			UserMessageID: dm.ID,
			BotMessageID:  msg.ID,
		}
		return
	}

	ctx.Content = strings.Join(parts[1:], " ")

	msg := respond(fmt.Sprintf("🔺Translate:\n❝ %s ❞\nto\n❝ %s ❞", st.Tweet.Text, ctx.Content))
	err = ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\u2705")
	if err != nil {
		fmt.Printf("Failed to add \u2705 reaction, %s\n", err)
	}
	err = ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\u274C")
	if err != nil {
		fmt.Printf("Failed to add \u274C reaction, %s\n", err)
	}
	config.TweetUpdates[msg.ID] = config.TweetUpdate{
		ChannelID:      msg.ChannelID,
		UserMessageID:  dm.ID,
		BotMessageID:   msg.ID,
		TweetMessageID: st.MessageID,
		Translation:    ctx.Content,
		Translator:     dm.Author.Username,
	}
}

func (m *Mux) CancelTweetUpdate(ds *discordgo.Session, tu config.TweetUpdate) {
	ds.ChannelMessageDelete(tu.ChannelID, tu.UserMessageID)
	ds.ChannelMessageDelete(tu.ChannelID, tu.BotMessageID)
}

func (m *Mux) DoTweetUpdate(ds *discordgo.Session, tu config.TweetUpdate) {
	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)
	stCol := db.C("synced_tweets")

	st := models.SyncedTweet{}
	err := stCol.Find(bson.M{"message_id": tu.TweetMessageID}).One(&st)
	if err != nil {
		ds.ChannelMessageSend(tu.ChannelID, fmt.Sprintf("Error updating tweet: %s", err))
		return
	}

	if !st.HumanTranslated {
		st.HumanTranslated = true
		st.Translators = []string{}
	}

	st.Translation = tu.Translation
	found := false
	for _, t := range st.Translators {
		if t == tu.Translator {
			found = true
		}
	}
	if !found {
		st.Translators = append(st.Translators, tu.Translator)
	}
	st.UpdatedAt = time.Now()

	err = stCol.Update(bson.M{"message_id": st.MessageID}, st)
	if err != nil {
		ds.ChannelMessageSend(tu.ChannelID, fmt.Sprintf("Error updating tweet: %s", err))
		return
	}

	embed := tweetsync.SyncedTweetToEmbed(st)

	_, err = ds.ChannelMessageEditEmbed(tu.ChannelID, st.MessageID, embed)
	if err != nil {
		ds.ChannelMessageSend(tu.ChannelID, fmt.Sprintf("Error updating tweet: %s", err))
		return
	}

	ds.ChannelMessageDelete(tu.ChannelID, tu.UserMessageID)
	ds.ChannelMessageDelete(tu.ChannelID, tu.BotMessageID)
}

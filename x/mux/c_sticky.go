package mux

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/w8kerr/delubot/models"
	"github.com/w8kerr/delubot/mongo"
)

func (m *Mux) Sticky(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	ctx.Content = strings.TrimPrefix(ctx.Content, "sticky")
	ctx.Content = strings.TrimSpace(ctx.Content)

	if ctx.Content == "" {
		respond("🔺Usage: -db sticky <the message you want to sticky>")
		return
	}

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)
	sCol := db.C("stickies")

	sticky := models.Sticky{}
	err := sCol.Find(bson.M{"channel_id": dm.ChannelID}).One(&sticky)
	if err != nil && err != mgo.ErrNotFound {
		respond(fmt.Sprintf("🔺Failed to check stickies: %s", err))
		return
	}
	if err == nil {
		ds.ChannelMessageDelete(sticky.ChannelID, sticky.MessageID)
		sticky.Text = ctx.Content
		sticky.Time = time.Now()
		sticky.AuthorName = dm.Author.Username
		sticky.AuthorAvatarURL = dm.Author.AvatarURL("")
		msg, err := ds.ChannelMessageSendEmbed(sticky.ChannelID, StickyEmbed(sticky))
		if err != nil {
			log.Printf("Error sending sticky message: %s", err)
			return
		}
		sticky.MessageID = msg.ID
		err = sCol.UpdateId(sticky.OID, sticky)
		if err != nil {
			log.Printf("Error updating sticky message: %s", err)
			return
		}
	} else {
		sticky = models.Sticky{
			OID:             bson.NewObjectId(),
			AuthorName:      dm.Author.Username,
			AuthorAvatarURL: dm.Author.AvatarURL(""),
			ChannelID:       dm.ChannelID,
			Text:            ctx.Content,
			Time:            time.Now(),
		}

		msg, err := ds.ChannelMessageSendEmbed(sticky.ChannelID, StickyEmbed(sticky))
		if err != nil {
			log.Printf("Error sending sticky message: %s", err)
			return
		}
		sticky.MessageID = msg.ID
		err = sCol.Insert(sticky)
		if err != nil {
			log.Printf("Error updating sticky message: %s", err)
			return
		}
	}
}

func (m *Mux) Unsticky(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)
	sCol := db.C("stickies")

	sticky := models.Sticky{}
	err := sCol.Find(bson.M{"channel_id": dm.ChannelID}).One(&sticky)
	if err != nil {
		respond(fmt.Sprintf("🔺Failed to unsticky message: %s", err))
		return
	}

	err = sCol.Remove(bson.M{"channel_id": dm.ChannelID})
	if err != nil {
		respond(fmt.Sprintf("🔺Failed to unsticky message: %s", err))
		return
	}
	respond("🔺Message unstickied")
}

func (m *Mux) EnsureSticky(db *mgo.Database, ds *discordgo.Session, dm *discordgo.Message) bool {
	if len(dm.Embeds) > 0 && dm.Embeds[0].Footer != nil && dm.Embeds[0].Footer.Text == "Sticky" {
		return true
	}

	sCol := db.C("stickies")

	sticky := models.Sticky{}
	err := sCol.Find(bson.M{"channel_id": dm.ChannelID}).One(&sticky)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("Failed to check for sticky: %s", err)
		return false
	}

	// Simply continue if the channel has no sticky
	if err == mgo.ErrNotFound {
		return false
	}

	fmt.Println("Processing message", dm.ID)
	fmt.Println("Sticky found with ID", sticky.MessageID)

	// Abort (and tell the rest of the bot's functions to abort as well) if the message is the sticky message
	if dm.ID == sticky.MessageID {
		fmt.Println("Message was sticky message, abort!")
		return true
	}

	// Otherwise, repost the sticky
	fmt.Println("Delete old sticky")
	ds.ChannelMessageDelete(sticky.ChannelID, sticky.MessageID)
	fmt.Println("Post new sticky")
	msg, err := ds.ChannelMessageSendEmbed(sticky.ChannelID, StickyEmbed(sticky))
	if err != nil {
		log.Printf("Error refreshing sticky message: %s", err)
		return false
	}
	sticky.MessageID = msg.ID
	fmt.Println("Update sticky record with new messageID", msg.ID)
	err = sCol.UpdateId(sticky.OID, sticky)
	if err != nil {
		log.Printf("Error updating sticky message: %s", err)
		return false
	}
	fmt.Println("Done")
	return false
}

func StickyEmbed(sticky models.Sticky) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color: 3066993,
		// Thumbnail: &discordgo.MessageEmbedThumbnail{
		// 	URL: sticky.AuthorAvatarURL,
		// },
		Author: &discordgo.MessageEmbedAuthor{
			Name: sticky.AuthorName,
		},
		Description: sticky.Text,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Sticky",
		},
		Timestamp: sticky.Time.Format(time.RFC3339),
	}
	return embed
}

package mux

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
)

func (m *Mux) Headpat(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	target := dm.Author
	if dm.MessageReference != nil {
		msg, err := ds.ChannelMessage(dm.MessageReference.ChannelID, dm.MessageReference.MessageID)
		if err != nil {
			log.Printf("Failed to get reply message: %s", err)
			respond("🔺FLAGRANT HEADPAT ERROR. COMPUTER OVER.")
		}
		target = msg.Author
	}

	veeMention := "<@!288848889174556682>"
	msg := fmt.Sprintf("_-Pats %s's head-_", target.Mention())
	emoji := config.Emoji("delupat")

	rand.Seed(time.Now().UnixNano())
	if rand.Intn(2) == 0 {
		msg = fmt.Sprintf("_-Pats %s's head instead-_", veeMention)
		emoji = config.Emoji("VeePat")
	} else {
		if dm.Author.Username == "default" {
			emoji = config.Emoji("defaultpat")
		}
		if dm.Author.Username == "Mirrored" {
			emoji = config.Emoji("mirroredpat")
		}
		if dm.Author.Username == "Kitsu 木狐" {
			msg = fmt.Sprintf("_-Gives %s a thumbs up-_", dm.Author.Mention())
			emoji = config.Emoji("okaytsu")
		}
		if dm.Author.Username == "EarthenSpire" {
			emoji = config.Emoji("stickpat")
		}
	}

	respond(msg)
	respond(emoji)
}

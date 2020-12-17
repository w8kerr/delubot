package mux

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
)

func (m *Mux) Headpat(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	veeMention := "<@!288848889174556682>"
	msg := fmt.Sprintf("_-Pats %s's head-_", dm.Author.Mention())

	rand.Seed(time.Now().UnixNano())
	if rand.Intn(2) == 0 {
		msg = fmt.Sprintf("_-Pats %s's head instead-_", veeMention)
	}

	respond(msg)
	respond(config.Emoji("delupat"))
}

package mux

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (m *Mux) Nickname(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	ctx.Content = strings.TrimPrefix(ctx.Content, "nickname")
	ctx.Content = strings.TrimSpace(ctx.Content)
	if ctx.Content == "" {
		respond("🔺Usage: -db nickname <name>")
		return
	}

	err := ds.GuildMemberNickname(dm.GuildID, "@me", ctx.Content)
	if err != nil {
		respond("🔺Failed to update nickname, " + err.Error())
		return
	}

	respond(fmt.Sprintf("🔺Nickname updated to \"%s\"!", ctx.Content))
}

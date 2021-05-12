package mux

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
)

var ThumbsUp = "\U0001F44D"

func (m *Mux) DoubleTL(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	prerespond := GetResponder(ds, dm)

	// emoji, err := ds.State.Emoji(dm.GuildID, "788243303816364062")
	// if err != nil {
	// 	prerespond(fmt.Sprintf("🔺No more eight ball I dropped it on the floor (" + err.Error() + ")"))
	// 	return
	// }

	ctx.Content = strings.TrimPrefix(ctx.Content, "doubletl")
	ctx.Content = strings.TrimSpace(ctx.Content)

	if ctx.Content == "disable" {
		ds.MessageReactionAdd(dm.ChannelID, dm.ID, ThumbsUp)
		config.SetDoubleTLEnabled(false)
		return
	}
	if ctx.Content == "enable" {
		ds.MessageReactionAdd(dm.ChannelID, dm.ID, ThumbsUp)
		config.SetDoubleTLEnabled(true)
		return
	}

	if ctx.Content == "" {
		prerespond("🔺Usage: -db doubletl <enable/disable>")
		return
	}
}

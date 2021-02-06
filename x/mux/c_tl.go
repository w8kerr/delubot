package mux

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/tl"
)

func (m *Mux) Translate(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	ctx.Content = strings.TrimPrefix(ctx.Content, "tl")
	ctx.Content = strings.TrimSpace(ctx.Content)

	if ctx.Content == "" {
		respond("🔺Usage: -db tl <text to translate>")
		return
	}

	translation, _, err := tl.DeepLTranslate(ctx.Content, deepl.LangEN)
	if err != nil {
		respond(fmt.Sprintf("🔺DeepL failed, %s", err))
		return
	}

	respond(fmt.Sprintf("🔺DeepL Translation:\n❝ %s ❞", translation))
}

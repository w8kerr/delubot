package mux

import (
	"github.com/bwmarrin/discordgo"
)

func (m *Mux) TestMsg(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	resp := "```mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"

	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"

	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"

	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"

	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu"
	resp += "mumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumumu```"

	respond(resp)
}

package mux

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/utils"
)

func (m *Mux) Config(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	resp := "Config!```"
	resp += "Grant roles: " + utils.PrintJSONStr(config.GrantRoles)
	resp += "\nModerator roles: " + utils.PrintJSONStr(config.ModeratorRoles)
	resp += "\nSync sheets: " + utils.PrintJSONStr(config.SyncSheets)
	resp += "\nSync enabled: " + utils.PrintJSONStr(config.SyncEnabled)
	resp += "\nTime format: " + utils.PrintJSONStr(config.TimeFormat)
	resp += "\nGoogle Credentials: Secret!"
	resp += "```"

	respond(resp)
}

func (m *Mux) RefreshConfig(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	err := config.LoadConfig()
	if err != nil {
		respond(fmt.Sprintf("Failed to load config, %s", err))
		return
	}

	respond("Updated config!")
}

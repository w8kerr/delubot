package mux

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
)

func (m *Mux) AlphaRole(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := func(msg string) {
		_, err := ds.ChannelMessageSend(dm.ChannelID, msg)
		if err != nil {
			fmt.Println(err)
		}
	}

	roles, err := ds.GuildRoles(dm.GuildID)
	if err != nil {
		respond(err.Error())
		return
	}

	alphaRoleID := config.AlphaRole(dm.GuildID)
	for _, role := range roles {
		if role.ID == alphaRoleID {
			resp := fmt.Sprintf("Alpha role: %s", role.Name)
			respond(resp)
			return
		}
	}

	respond("No Alpha role is configured")
}

func (m *Mux) WhaleRole(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := func(msg string) {
		_, err := ds.ChannelMessageSend(dm.ChannelID, msg)
		if err != nil {
			fmt.Println(err)
		}
	}

	roles, err := ds.GuildRoles(dm.GuildID)
	if err != nil {
		respond(err.Error())
		return
	}

	alphaRoleID := config.WhaleRole(dm.GuildID)
	for _, role := range roles {
		if role.ID == alphaRoleID {
			resp := fmt.Sprintf("Whale role: %s", role.Name)
			respond(resp)
			return
		}
	}

	respond("No Whale role is configured")
}

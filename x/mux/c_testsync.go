package mux

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/sheetsync"
)

func (m *Mux) TestSync(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	sheetID := config.SyncSheet(dm.GuildID)
	if sheetID == "" {
		respond("Could test role sync, no sync Sheet defined")
		return
	}

	alphaRole := config.AlphaRole(dm.GuildID)
	if sheetID == "" {
		respond("Could test role sync, no Alpha role defined")
		return
	}
	whaleRole := config.WhaleRole(dm.GuildID)
	if sheetID == "" {
		respond("Could test role sync, no Whale role defined")
		return
	}

	svc, err := sheetsync.GetService()
	if err != nil {
		log.Printf("Couldn't create Sheet service, %s", err)
		return
	}

	ga := func(member *discordgo.Member, entry sheetsync.RoleRow, errors *[]sheetsync.RoleRow, updated *bool, failed *bool) {
		handle := member.User.Username + "#" + member.User.Discriminator
		if !sheetsync.HasRole(member, alphaRole) {
			log.Println("Give Alpha role to", handle)
			*updated = true
		} else {
			log.Println(handle, " already had Alpha role")
		}
	}
	gw := func(member *discordgo.Member, entry sheetsync.RoleRow, errors *[]sheetsync.RoleRow, updated *bool, failed *bool) {
		handle := member.User.Username + "#" + member.User.Discriminator
		if !sheetsync.HasRole(member, whaleRole) {
			log.Println("Give Alpha role to", handle)
			*updated = true
		} else {
			log.Println(handle, " already had Whale role")
		}
	}
	ra := func(member *discordgo.Member, entry sheetsync.RoleRow, errors *[]sheetsync.RoleRow, updated *bool, failed *bool) {
		handle := member.User.Username + "#" + member.User.Discriminator
		if sheetsync.HasRole(member, alphaRole) {
			log.Println("Remove Alpha role from", handle)
			*updated = true
		} else {
			log.Println(handle, " already didn't have Alpha role")
		}
	}
	rw := func(member *discordgo.Member, entry sheetsync.RoleRow, errors *[]sheetsync.RoleRow, updated *bool, failed *bool) {
		handle := member.User.Username + "#" + member.User.Discriminator
		if sheetsync.HasRole(member, whaleRole) {
			log.Println("Remove Whale role from", handle)
			*updated = true
		} else {
			log.Println(handle, " already had Whale role")
		}
	}

	report := func(gaveAlpha []sheetsync.RoleRow, gaveWhale []sheetsync.RoleRow, removedRoles []sheetsync.RoleRow, wasBanned []sheetsync.RoleRow, errors []sheetsync.RoleRow) {
		resp := "Here's what would sync!```"
		resp += fmt.Sprintf("Grant Alpha role to %d members", len(gaveAlpha))
		resp += fmt.Sprintf("\nGrant Whale role to %d members", len(gaveWhale))
		resp += fmt.Sprintf("\nRemove roles from %d members (%d because of bans)", len(removedRoles), len(wasBanned))
		resp += "```"

		respond(resp)
	}

	sheetsync.DoSyncGuild(svc, dm.GuildID, ga, gw, ra, rw, report, false)
}
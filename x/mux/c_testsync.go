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
	if alphaRole == "" {
		respond("Could test role sync, no Alpha role defined")
		return
	}
	specialRole := config.SpecialRole(dm.GuildID)
	if specialRole == "" {
		respond("Could test role sync, no Whale role defined")
		return
	}
	whaleRole := config.WhaleRole(dm.GuildID)
	if whaleRole == "" {
		respond("Could test role sync, no Whale role defined")
		return
	}
	fanboxRole := config.FanboxRole(dm.GuildID)
	if fanboxRole == "" {
		respond("Could test role sync, no Fanbox role defined")
		return
	}

	svc, err := sheetsync.GetService()
	if err != nil {
		log.Printf("Couldn't create Sheet service, %s", err)
		return
	}

	page, _, _, err := sheetsync.GetCurrentPage(svc, sheetID)
	if err != nil {
		log.Printf("%s - Couldn't get the current page, %s", dm.GuildID, err)
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
	gs := func(member *discordgo.Member, entry sheetsync.RoleRow, errors *[]sheetsync.RoleRow, updated *bool, failed *bool) {
		handle := member.User.Username + "#" + member.User.Discriminator
		if !sheetsync.HasRole(member, specialRole) {
			log.Println("Give Special role to", handle)
			*updated = true
		} else {
			log.Println(handle, " already had Special role")
		}
	}
	gw := func(member *discordgo.Member, entry sheetsync.RoleRow, errors *[]sheetsync.RoleRow, updated *bool, failed *bool) {
		handle := member.User.Username + "#" + member.User.Discriminator
		if !sheetsync.HasRole(member, whaleRole) {
			log.Println("Give Whale role to", handle)
			*updated = true
		} else {
			log.Println(handle, " already had Whale role")
		}
	}
	gf := func(member *discordgo.Member, entry sheetsync.RoleRow, errors *[]sheetsync.RoleRow, updated *bool, failed *bool) {
		handle := member.User.Username + "#" + member.User.Discriminator
		if !sheetsync.HasRole(member, fanboxRole) {
			log.Println("Give Fanbox role to", handle)
			*updated = true
		} else {
			log.Println(handle, " already had Fanbox role")
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
	rs := func(member *discordgo.Member, entry sheetsync.RoleRow, errors *[]sheetsync.RoleRow, updated *bool, failed *bool) {
		handle := member.User.Username + "#" + member.User.Discriminator
		if sheetsync.HasRole(member, specialRole) {
			log.Println("Remove Special role from", handle)
			*updated = true
		} else {
			log.Println(handle, " already didn't have Special role")
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
	rf := func(member *discordgo.Member, entry sheetsync.RoleRow, errors *[]sheetsync.RoleRow, updated *bool, failed *bool) {
		handle := member.User.Username + "#" + member.User.Discriminator
		if sheetsync.HasRole(member, fanboxRole) {
			log.Println("Remove Fanbox role from", handle)
			*updated = true
		} else {
			log.Println(handle, " already had Fanbox role")
		}
	}

	report := func(gaveAlpha []sheetsync.RoleRow, gaveSpecial []sheetsync.RoleRow, gaveWhale []sheetsync.RoleRow, gaveFanbox []sheetsync.RoleRow, removedRoles []sheetsync.RoleRow, wasBanned []sheetsync.RoleRow, errors []sheetsync.RoleRow) {
		resp := "Here's what would sync!```"
		resp += fmt.Sprintf("Grant Alpha role to %d members", len(gaveAlpha))
		resp += fmt.Sprintf("\nGrant Special role to %d members", len(gaveSpecial))
		resp += fmt.Sprintf("\nGrant Whale role to %d members", len(gaveWhale))
		resp += fmt.Sprintf("\nGrant Fanbox role to %d members", len(gaveFanbox))
		resp += fmt.Sprintf("\nRemove roles from %d members (%d because of bans)", len(removedRoles), len(wasBanned))
		resp += "```"

		respond(resp)
	}

	//sheetsync.DoSyncGuild(svc, dm.GuildID, sheetID, page, ga, gs, gw, gf, ra, rs, rw, rf, report, false)
	sheetsync.DoSyncGuildV2(svc, dm.GuildID, sheetID, page, ga, gs, gw, gf, ra, rs, rw, rf, report, false)
}

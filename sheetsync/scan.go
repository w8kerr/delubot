package sheetsync

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/utils"
	"google.golang.org/api/sheets/v4"
)

func Sweeper() {
	sleepDuration := 60 * time.Second
	for {
		time.Sleep(sleepDuration)
		Scan()
	}
}

func Scan() {
	// log.Println("Starting discord roles sync", time.Now())
	syncGuilds := make(map[string]bool)
	for guildID, doSync := range config.RoleGrantEnabled {
		if doSync {
			syncGuilds[guildID] = true
		}
	}
	for guildID, doSync := range config.RoleRemoveEnabled {
		if doSync {
			syncGuilds[guildID] = true
		}
	}

	if len(syncGuilds) == 0 {
		log.Printf("No guilds to sync, done for now")
		return
	}

	svc, err := GetService()
	if err != nil {
		log.Printf("Couldn't create Sheet service, %s", err)
		return
	}

	for guildID := range syncGuilds {
		sheetID := config.SyncSheet(guildID)
		if sheetID == "" {
			log.Println("Skipped sync for", guildID, ", no sync Sheet in config")
			continue
		}

		alphaRole := config.AlphaRole(guildID)
		if alphaRole == "" {
			log.Println("Skipped sync for", guildID, ", no Alpha role in config")
			continue
		}
		specialRole := config.SpecialRole(guildID)
		if specialRole == "" {
			log.Println("Skipped sync for", guildID, ", no Special role in config")
			continue
		}
		whaleRole := config.WhaleRole(guildID)
		if whaleRole == "" {
			log.Println("Skipped sync for", guildID, ", no Whale role in config")
			continue
		}

		SyncGuild(svc, guildID)
	}
}

func SyncGuild(svc *sheets.Service, guildID string) {
	alphaRole := config.AlphaRole(guildID)
	specialRole := config.SpecialRole(guildID)
	whaleRole := config.WhaleRole(guildID)
	formerRole := config.FormerRole(guildID)

	roleGrant := config.RoleGrantIsEnabled(guildID)
	roleRemove := config.RoleRemoveIsEnabled(guildID)

	sheetID := config.SyncSheet(guildID)
	page, doRemove, err := GetCurrentPage(svc, sheetID)
	if err != nil {
		log.Printf("%s - Couldn't get the current page, %s", guildID, err)
		return
	}

	roleRemove = roleRemove && doRemove

	ensureAlpha := func(member *discordgo.Member, entry RoleRow, errors *[]RoleRow, updated *bool, failed *bool) {
		if !roleGrant {
			return
		}
		if !HasRole(member, alphaRole) {
			err := Session.GuildMemberRoleAdd(guildID, member.User.ID, alphaRole)
			if err != nil {
				*errors = append(*errors, entry)
				*failed = true
			} else {
				*updated = true
				if HasRole(member, formerRole) {
					err = Session.GuildMemberRoleRemove(guildID, member.User.ID, formerRole)
					*errors = append(*errors, entry)
				}
			}
		}
	}
	ensureSpecial := func(member *discordgo.Member, entry RoleRow, errors *[]RoleRow, updated *bool, failed *bool) {
		if !roleGrant {
			return
		}
		if !HasRole(member, specialRole) {
			err := Session.GuildMemberRoleAdd(guildID, member.User.ID, specialRole)
			if err != nil {
				*errors = append(*errors, entry)
				*failed = true
			} else {
				*updated = true
			}
		}
	}
	ensureWhale := func(member *discordgo.Member, entry RoleRow, errors *[]RoleRow, updated *bool, failed *bool) {
		if !roleGrant {
			return
		}
		if !HasRole(member, whaleRole) {
			err := Session.GuildMemberRoleAdd(guildID, member.User.ID, whaleRole)
			if err != nil {
				*errors = append(*errors, entry)
				*failed = true
			} else {
				*updated = true
			}
		}
	}
	ensureNoAlpha := func(member *discordgo.Member, entry RoleRow, errors *[]RoleRow, updated *bool, failed *bool) {
		if !roleRemove {
			return
		}
		if HasRole(member, alphaRole) {
			err := Session.GuildMemberRoleRemove(guildID, member.User.ID, alphaRole)
			if err != nil {
				*errors = append(*errors, entry)
				*failed = true
			} else {
				*updated = true
				if !HasRole(member, formerRole) {
					err = Session.GuildMemberRoleAdd(guildID, member.User.ID, formerRole)
					*errors = append(*errors, entry)
				}
			}
		}
	}
	ensureNoSpecial := func(member *discordgo.Member, entry RoleRow, errors *[]RoleRow, updated *bool, failed *bool) {
		if !roleRemove {
			return
		}
		if HasRole(member, specialRole) {
			err := Session.GuildMemberRoleRemove(guildID, member.User.ID, specialRole)
			if err != nil {
				*errors = append(*errors, entry)
				*failed = true
			} else {
				*updated = true
			}
		}
	}
	ensureNoWhale := func(member *discordgo.Member, entry RoleRow, errors *[]RoleRow, updated *bool, failed *bool) {
		if !roleRemove {
			return
		}
		if HasRole(member, whaleRole) {
			err := Session.GuildMemberRoleRemove(guildID, member.User.ID, whaleRole)
			if err != nil {
				*errors = append(*errors, entry)
				*failed = true
			} else {
				*updated = true
			}
		}
	}

	report := func(gaveAlpha []RoleRow, gaveSpecial []RoleRow, gaveWhale []RoleRow, removedRoles []RoleRow, wasBanned []RoleRow, errors []RoleRow) {
		if len(gaveAlpha) > 0 {
			log.Printf("Granted Alpha role to %d members", len(gaveAlpha))
		}
		if len(gaveSpecial) > 0 {
			log.Printf("Granted Special role to %d members", len(gaveSpecial))
		}
		if len(gaveWhale) > 0 {
			log.Printf("Granted Whale role to %d members", len(gaveWhale))
		}
		if len(removedRoles) > 0 {
			log.Printf("Removed roles from %d members (%d because of bans)", len(removedRoles), len(wasBanned))
		}

		if len(errors) > 0 {
			for _, err := range errors {
				log.Println("Failed to process roles for", err.Handle())
			}
		}
	}

	DoSyncGuild(svc, guildID, sheetID, page, ensureAlpha, ensureSpecial, ensureWhale, ensureNoAlpha, ensureNoSpecial, ensureNoWhale, report, true)
}

func DoSyncGuild(svc *sheets.Service, guildID string, sheetID string, page *sheets.Sheet,
	ensureAlpha func(*discordgo.Member, RoleRow, *[]RoleRow, *bool, *bool),
	ensureSpecial func(*discordgo.Member, RoleRow, *[]RoleRow, *bool, *bool),
	ensureWhale func(*discordgo.Member, RoleRow, *[]RoleRow, *bool, *bool),
	ensureNoAlpha func(*discordgo.Member, RoleRow, *[]RoleRow, *bool, *bool),
	ensureNoSpecial func(*discordgo.Member, RoleRow, *[]RoleRow, *bool, *bool),
	ensureNoWhale func(*discordgo.Member, RoleRow, *[]RoleRow, *bool, *bool),
	report func([]RoleRow, []RoleRow, []RoleRow, []RoleRow, []RoleRow, []RoleRow), doFormat bool) {

	entries, err := ReadAllAutomatic(svc, sheetID, page)
	if err != nil {
		log.Printf("%s - Failed to read automatic Sheet rows, %s", guildID, err)
		return
	}
	// log.Printf("%s - Got %d automatic entries", guildID, len(entries))

	manualEntries, err := ReadAllManual(svc, sheetID, page)
	if err != nil {
		log.Printf("%s - Failed to read manual Sheet rows, %s", guildID, err)
		return
	}
	// log.Printf("%s - Got %d manual entries", guildID, len(manualEntries))

	entries = append(entries, manualEntries...)
	entryMap := MapRows(entries)

	excluded, err := ReadAllExclude(svc, sheetID, page)
	if err != nil {
		log.Printf("%s - Failed to read excluded Sheet rows, %s", guildID, err)
		return
	}
	banMap := MapRows(excluded)

	members, err := utils.GetAllMembers(Session, guildID)
	if err != nil {
		log.Printf("%s - Failed to get guild members, %s", guildID, err)
		return
	}
	// log.Printf("%s - Processing %d members", guildID, len(members))

	formatReqs := []*sheets.Request{}
	gaveAlpha := []RoleRow{}
	gaveSpecial := []RoleRow{}
	gaveWhale := []RoleRow{}
	removedRoles := []RoleRow{}
	wasBanned := []RoleRow{}
	errors := []RoleRow{}

	for _, member := range members {
		handle := member.User.Username + "#" + member.User.Discriminator
		entry, hasEntry := entryMap[member.User.ID]

		if hasEntry {
			ban, hasBan := banMap[member.User.ID]
			if hasBan {
				updated := false
				failed := false
				ensureNoAlpha(member, entry, &errors, &updated, &failed)
				ensureNoSpecial(member, entry, &errors, &updated, &failed)
				ensureNoWhale(member, entry, &errors, &updated, &failed)

				if updated {
					formatReqs = append(formatReqs, entry.ColorRequest(RedHighlight))
					formatReqs = append(formatReqs, ban.ColorRequest(RedHighlight))
				}
				if updated && !failed {
					removedRoles = append(removedRoles, entry)
					wasBanned = append(wasBanned, ban)
				}
			} else {
				updated := false
				failed := false
				var color sheets.Color
				if entry.Plan >= 500 {
					ensureAlpha(member, entry, &errors, &updated, &failed)
					color = GreenHighlight
				}
				if entry.Plan >= 1500 {
					ensureSpecial(member, entry, &errors, &updated, &failed)
					color = BlueHighlight
				}
				if entry.Plan >= 10000 {
					ensureWhale(member, entry, &errors, &updated, &failed)
					color = YellowHighlight
				}
				if entry.Plan < 1500 {
					ensureNoSpecial(member, entry, &errors, &updated, &failed)
				}
				if entry.Plan < 10000 {
					ensureNoWhale(member, entry, &errors, &updated, &failed)
				}

				if updated {
					formatReqs = append(formatReqs, entry.ColorRequest(color))
				}
				if updated && !failed {
					if entry.Plan >= 500 {
						gaveAlpha = append(gaveAlpha, entry)
					}
					if entry.Plan >= 1500 {
						gaveSpecial = append(gaveSpecial, entry)
					}
					if entry.Plan >= 10000 {
						gaveWhale = append(gaveWhale, entry)
					}
				}
			}

			if entry.Handle() != handle {
				err = UpdateHandle(svc, sheetID, page, entry, handle)
				if err != nil {
					log.Printf("ERROR: Failed to update handle (%s)\n", member.User.ID)
				} else {
					log.Printf("Update handle from '%s' to '%s' (%s)\n", entry.Handle(), handle, member.User.ID)
				}
			}
		} else {
			updated := false
			failed := false
			ensureNoAlpha(member, entry, &errors, &updated, &failed)
			ensureNoSpecial(member, entry, &errors, &updated, &failed)
			ensureNoWhale(member, entry, &errors, &updated, &failed)

			if updated && !failed {
				removedRoles = append(removedRoles, RoleRow{
					Username:      member.User.Username,
					Discriminator: member.User.Discriminator,
				})
			}
		}
	}

	if doFormat && len(formatReqs) > 0 {
		err = UpdateFormatting(svc, sheetID, formatReqs)
		if err != nil {
			log.Printf("%s - Failed to update formatting, %s", guildID, err)
		}
	}

	report(gaveAlpha, gaveSpecial, gaveWhale, removedRoles, wasBanned, errors)
}

func HasRole(member *discordgo.Member, roleID string) bool {
	for _, rID := range member.Roles {
		if rID == roleID {
			return true
		}
	}

	return false
}

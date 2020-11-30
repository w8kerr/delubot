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
	log.Println("Starting discord roles sync", time.Now())
	syncGuilds := []string{}
	for guildID, doSync := range config.SyncEnabled {
		if !doSync {
			log.Println("Skipped sync for", guildID, ", disabled in config")
			continue
		}

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
		whaleRole := config.WhaleRole(guildID)
		if whaleRole == "" {
			log.Println("Skipped sync for", guildID, ", no Whale role in config")
			continue
		}

		syncGuilds = append(syncGuilds, guildID)
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

	for _, guildID := range syncGuilds {
		SyncGuild(svc, guildID)
	}
}

func SyncGuild(svc *sheets.Service, guildID string) {
	alphaRole := config.AlphaRole(guildID)
	whaleRole := config.WhaleRole(guildID)

	ensureAlpha := func(member *discordgo.Member, entry RoleRow, errors *[]RoleRow, updated *bool, failed *bool) {
		if !HasRole(member, alphaRole) {
			err := Session.GuildMemberRoleAdd(guildID, member.User.ID, alphaRole)
			if err != nil {
				*errors = append(*errors, entry)
				*failed = true
			} else {
				*updated = true
			}
		}
	}
	ensureWhale := func(member *discordgo.Member, entry RoleRow, errors *[]RoleRow, updated *bool, failed *bool) {
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
		if HasRole(member, alphaRole) {
			err := Session.GuildMemberRoleRemove(guildID, member.User.ID, alphaRole)
			if err != nil {
				*errors = append(*errors, entry)
				*failed = true
			} else {
				*updated = true
			}
		}
	}
	ensureNoWhale := func(member *discordgo.Member, entry RoleRow, errors *[]RoleRow, updated *bool, failed *bool) {
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

	report := func(gaveAlpha []RoleRow, gaveWhale []RoleRow, removedRoles []RoleRow, wasBanned []RoleRow, errors []RoleRow) {
		log.Printf("Granted Alpha role to %d members", len(gaveAlpha))
		log.Printf("Granted Whale role to %d members", len(gaveWhale))
		log.Printf("Removed roles from %d members (%d because of bans)", len(removedRoles), len(wasBanned))

		if len(errors) > 0 {
			for _, err := range errors {
				log.Println("Failed to process roles for", err.Handle())
			}
		}
	}

	DoSyncGuild(svc, guildID, ensureAlpha, ensureWhale, ensureNoAlpha, ensureNoWhale, report, true)
}

func DoSyncGuild(svc *sheets.Service, guildID string,
	ensureAlpha func(*discordgo.Member, RoleRow, *[]RoleRow, *bool, *bool),
	ensureWhale func(*discordgo.Member, RoleRow, *[]RoleRow, *bool, *bool),
	ensureNoAlpha func(*discordgo.Member, RoleRow, *[]RoleRow, *bool, *bool),
	ensureNoWhale func(*discordgo.Member, RoleRow, *[]RoleRow, *bool, *bool),
	report func([]RoleRow, []RoleRow, []RoleRow, []RoleRow, []RoleRow), doFormat bool) {

	sheetID := config.SyncSheet(guildID)
	page, doRemove, err := GetCurrentPage(svc, sheetID)
	if err != nil {
		log.Printf("%s - Couldn't get the current page, %s", guildID, err)
		return
	}

	entries, err := ReadAllAutomatic(svc, sheetID, page)
	if err != nil {
		log.Printf("%s - Failed to read automatic Sheet rows, %s", guildID, err)
		return
	}
	log.Printf("%s - Got %d automatic entries", guildID, len(entries))

	manualEntries, err := ReadAllManual(svc, sheetID, page)
	if err != nil {
		log.Printf("%s - Failed to read manual Sheet rows, %s", guildID, err)
		return
	}
	log.Printf("%s - Got %d manual entries", guildID, len(manualEntries))

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
	log.Printf("%s - Processing %d members", guildID, len(members))

	formatReqs := []*sheets.Request{}
	gaveAlpha := []RoleRow{}
	gaveWhale := []RoleRow{}
	removedRoles := []RoleRow{}
	wasBanned := []RoleRow{}
	errors := []RoleRow{}

	for _, member := range members {
		handle := member.User.Username + "#" + member.User.Discriminator
		entry, hasEntry := entryMap[handle]
		if hasEntry {
			ban, hasBan := banMap[handle]
			if hasBan {
				updated := false
				failed := false
				ensureNoAlpha(member, entry, &errors, &updated, &failed)
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
				color := GreenHighlight
				if entry.Plan == "10000" {
					ensureAlpha(member, entry, &errors, &updated, &failed)
					ensureWhale(member, entry, &errors, &updated, &failed)
					color = YellowHighlight
				} else {
					ensureAlpha(member, entry, &errors, &updated, &failed)
				}

				if updated {
					formatReqs = append(formatReqs, entry.ColorRequest(color))
				}
				if updated && !failed {
					if entry.Plan == "10000" {
						gaveWhale = append(gaveWhale, entry)
					} else {
						gaveAlpha = append(gaveAlpha, entry)
					}
				}
			}
		} else if doRemove {
			updated := false
			failed := false
			ensureNoAlpha(member, entry, &errors, &updated, &failed)
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

	report(gaveAlpha, gaveWhale, removedRoles, wasBanned, errors)
}

func HasRole(member *discordgo.Member, roleID string) bool {
	for _, rID := range member.Roles {
		if rID == roleID {
			return true
		}
	}

	return false
}

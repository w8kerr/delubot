package mux

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/sheetsync"
)

func (m *Mux) SyncSheet(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	ctx.Content = strings.TrimPrefix(ctx.Content, "syncsheet")
	ctx.Content = strings.TrimSpace(ctx.Content)
	if ctx.Content == "clear" {
		err := config.SetSyncSheet(dm.GuildID, "")
		if err != nil {
			respond(fmt.Sprintf("Failed to clear sync Sheet URL, %s", err))
			return
		}

		respond("Sync Sheet URL cleared")
		return
	} else if ctx.Content != "" {
		if !sheetsync.HasAccess(ctx.Content) {
			resp := fmt.Sprintf("Could not access the sheet ID `%s`\n", ctx.Content)
			resp += "Make sure the sheet is shared with the user `server@delutayaclub.iam.gserviceaccount.com`"
			respond(resp)
			return
		}

		err := config.SetSyncSheet(dm.GuildID, ctx.Content)
		if err != nil {
			respond(fmt.Sprintf("Failed to set sync Sheet, %s", err))
			return
		}

		respond(fmt.Sprintf("Sync Sheet set to `%s`", ctx.Content))
		return
	}

	syncSheet := config.SyncSheet(dm.GuildID)
	if syncSheet != "" {
		respond(fmt.Sprintf("Sync Sheet: `%s`", syncSheet))
		return
	}

	respond("No sync Sheet is configured")
}

func (m *Mux) Sync(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	enabled := config.SyncIsEnabled(dm.GuildID)

	ctx.Content = strings.TrimPrefix(ctx.Content, "sync")
	ctx.Content = strings.TrimSpace(ctx.Content)
	if ctx.Content == "enable" {
		if enabled {
			respond("Role sync is already enabled!")
			return
		}

		sheetID := config.SyncSheet(dm.GuildID)
		if sheetID == "" {
			respond("Can't enable role sync, no sync Sheet is defined")
			return
		}
		canAccess := sheetsync.HasAccess(sheetID)
		if !canAccess {
			respond(fmt.Sprintf("Can't enable role sync, the current sync Sheet `%s` could not be accessed", sheetID))
			return
		}

		err := config.SetSyncEnabled(dm.GuildID, true)
		if err != nil {
			respond(fmt.Sprintf("Failed to enable syncing, %s", err))
			return
		}

		respond("Role sync enabled!")
		return
	} else if ctx.Content == "disable" {
		if !enabled {
			respond("Role sync is already disabled!")
			return
		}

		err := config.SetSyncEnabled(dm.GuildID, false)
		if err != nil {
			respond(fmt.Sprintf("Failed to disable syncing, %s", err))
			return
		}

		respond("Role sync disabled!")
		return
	}

	if enabled {
		respond("Role sync is currently enabled!")
		return
	}
	respond("Role sync is currently disabled!")
}

func (m *Mux) AlphaRole(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	roles, err := ds.GuildRoles(dm.GuildID)
	if err != nil {
		respond(err.Error())
		return
	}

	ctx.Content = strings.TrimPrefix(ctx.Content, "alpharole")
	ctx.Content = strings.TrimSpace(ctx.Content)
	if ctx.Content == "clear" {
		err = config.SetAlphaRole(dm.GuildID, "")
		if err != nil {
			respond(fmt.Sprintf("Failed to clear alpha role, %s", err))
			return
		}

		respond("Alpha role cleared")
		return
	} else if ctx.Content != "" {
		roleID := ""
		roleName := ""
		for _, role := range roles {
			if role.ID == ctx.Content || role.Name == ctx.Content {
				roleID = role.ID
				roleName = role.Name
				break
			}
		}

		if roleID == "" {
			respond(fmt.Sprintf("No role found matching ID or Name '%s'", ctx.Content))
			return
		}

		err = config.SetAlphaRole(dm.GuildID, roleID)
		if err != nil {
			respond(fmt.Sprintf("Failed to set alpha role, %s", err))
			return
		}

		respond(fmt.Sprintf("Alpha role set to %s (`%s`)", roleName, roleID))
		return
	}

	alphaRoleID := config.AlphaRole(dm.GuildID)
	for _, role := range roles {
		if role.ID == alphaRoleID {
			resp := fmt.Sprintf("Alpha role: %s (`%s`)", role.Name, role.ID)
			respond(resp)
			return
		}
	}

	respond("No Alpha role is configured")
}

func (m *Mux) WhaleRole(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	roles, err := ds.GuildRoles(dm.GuildID)
	if err != nil {
		respond(err.Error())
		return
	}

	ctx.Content = strings.TrimPrefix(ctx.Content, "whalerole")
	ctx.Content = strings.TrimSpace(ctx.Content)
	if ctx.Content == "clear" {
		err = config.SetWhaleRole(dm.GuildID, "")
		if err != nil {
			respond(fmt.Sprintf("Failed to clear whale role, %s", err))
			return
		}

		respond("Whale role cleared")
		return
	} else if ctx.Content != "" {
		roleID := ""
		roleName := ""
		for _, role := range roles {
			if role.ID == ctx.Content || role.Name == ctx.Content {
				roleID = role.ID
				roleName = role.Name
				break
			}
		}

		if roleID == "" {
			respond(fmt.Sprintf("No role found matching ID or Name `%s`", ctx.Content))
			return
		}

		err = config.SetWhaleRole(dm.GuildID, roleID)
		if err != nil {
			respond(fmt.Sprintf("Failed to set whale role, %s", err))
			return
		}

		respond(fmt.Sprintf("Whale role set to %s (`%s`)", roleName, roleID))
		return
	}

	alphaRoleID := config.WhaleRole(dm.GuildID)
	for _, role := range roles {
		if role.ID == alphaRoleID {
			resp := fmt.Sprintf("Whale role: %s (`%s`)", role.Name, role.ID)
			respond(resp)
			return
		}
	}

	respond("No Whale role is configured")
}

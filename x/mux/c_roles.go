package mux

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/sheetsync"
)

func (m *Mux) SyncSheet(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	prerespond := GetResponder(ds, dm)
	msg := prerespond("Processing...")
	respond := GetEditor(ds, msg)

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
	if syncSheet == "" {
		respond("No sync Sheet is configured")
		return
	}

	svc, err := sheetsync.GetService()
	if err != nil {
		respond("Couldn't create Sheet service, " + err.Error())
		return
	}

	sheet, grantTime, removeTime, endTime, err := sheetsync.DoGetCurrentPage(svc, syncSheet)
	if err != nil {
		resp := fmt.Sprintf("Could not access the sheet ID `%s`\n", ctx.Content)
		resp += "Error: `" + err.Error() + "`"
		respond(resp)
		return
	}

	resp := fmt.Sprintf("```Sync from Google Sheet: %s", syncSheet)
	resp += fmt.Sprintf("\nCurrent month's page:   %s", sheet.Properties.Title)
	resp += fmt.Sprintf("\nStart granting roles:   %s", config.PrintTime(grantTime))
	resp += fmt.Sprintf("\nStart removing roles:   %s", config.PrintTime(removeTime))
	resp += fmt.Sprintf("\nEnd of sync:            %s", config.PrintTime(endTime))
	if config.RoleGrantIsEnabled(dm.GuildID) {
		resp += "\nRole granting:          Enabled"
	} else {
		resp += "\nRole granting:          Disabled"
	}
	if config.RoleRemoveIsEnabled(dm.GuildID) {
		resp += "\nRole removal:           Enabled"
	} else {
		resp += "\nRole removal:           Disabled"
	}
	resp += "```"

	respond(resp)
}

func (m *Mux) RoleGrant(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	enabled := config.RoleGrantIsEnabled(dm.GuildID)

	ctx.Content = strings.TrimPrefix(ctx.Content, "rolegrant")
	ctx.Content = strings.TrimSpace(ctx.Content)
	if ctx.Content == "enable" {
		if enabled {
			respond("Role granting is already enabled!")
			return
		}

		sheetID := config.SyncSheet(dm.GuildID)
		if sheetID == "" {
			respond("Can't enable role granting, no sync Sheet is defined")
			return
		}
		canAccess := sheetsync.HasAccess(sheetID)
		if !canAccess {
			respond(fmt.Sprintf("Can't enable role granting, the current sync Sheet `%s` could not be accessed", sheetID))
			return
		}

		err := config.SetRoleGrantEnabled(dm.GuildID, true)
		if err != nil {
			respond(fmt.Sprintf("Failed to enable role granting, %s", err))
			return
		}

		respond("Role granting enabled!")
		return
	} else if ctx.Content == "disable" {
		if !enabled {
			respond("Role granting is already disabled!")
			return
		}

		err := config.SetRoleGrantEnabled(dm.GuildID, false)
		if err != nil {
			respond(fmt.Sprintf("Failed to disable granting, %s", err))
			return
		}

		respond("Role granting disabled!")
		return
	}

	if enabled {
		respond("Role granting is currently enabled!")
		return
	}
	respond("Role granting is currently disabled!")
}

func (m *Mux) RoleRemove(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	enabled := config.RoleRemoveIsEnabled(dm.GuildID)

	ctx.Content = strings.TrimPrefix(ctx.Content, "roleremove")
	ctx.Content = strings.TrimSpace(ctx.Content)
	if ctx.Content == "enable" {
		if enabled {
			respond("Role removal is already enabled!")
			return
		}

		sheetID := config.SyncSheet(dm.GuildID)
		if sheetID == "" {
			respond("Can't enable role removal, no sync Sheet is defined")
			return
		}
		canAccess := sheetsync.HasAccess(sheetID)
		if !canAccess {
			respond(fmt.Sprintf("Can't enable role removal, the current sync Sheet `%s` could not be accessed", sheetID))
			return
		}

		err := config.SetRoleRemoveEnabled(dm.GuildID, true)
		if err != nil {
			respond(fmt.Sprintf("Failed to enable role removal, %s", err))
			return
		}

		respond("Role removal enabled!")
		return
	} else if ctx.Content == "disable" {
		if !enabled {
			respond("Role removal is already disabled!")
			return
		}

		err := config.SetRoleRemoveEnabled(dm.GuildID, false)
		if err != nil {
			respond(fmt.Sprintf("Failed to disable removal, %s", err))
			return
		}

		respond("Role removal disabled!")
		return
	}

	if enabled {
		respond("Role removal is currently enabled!")
		return
	}
	respond("Role removal is currently disabled!")
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

func (m *Mux) SpecialRole(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	roles, err := ds.GuildRoles(dm.GuildID)
	if err != nil {
		respond(err.Error())
		return
	}

	ctx.Content = strings.TrimPrefix(ctx.Content, "specialrole")
	ctx.Content = strings.TrimSpace(ctx.Content)
	if ctx.Content == "clear" {
		err = config.SetSpecialRole(dm.GuildID, "")
		if err != nil {
			respond(fmt.Sprintf("Failed to clear special role, %s", err))
			return
		}

		respond("Special role cleared")
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

		err = config.SetSpecialRole(dm.GuildID, roleID)
		if err != nil {
			respond(fmt.Sprintf("Failed to set special role, %s", err))
			return
		}

		respond(fmt.Sprintf("Special role set to %s (`%s`)", roleName, roleID))
		return
	}

	specialRoleID := config.SpecialRole(dm.GuildID)
	for _, role := range roles {
		if role.ID == specialRoleID {
			resp := fmt.Sprintf("Special role: %s (`%s`)", role.Name, role.ID)
			respond(resp)
			return
		}
	}

	respond("No Special role is configured")
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

func (m *Mux) FormerRole(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	roles, err := ds.GuildRoles(dm.GuildID)
	if err != nil {
		respond(err.Error())
		return
	}

	ctx.Content = strings.TrimPrefix(ctx.Content, "formerrole")
	ctx.Content = strings.TrimSpace(ctx.Content)
	if ctx.Content == "clear" {
		err = config.SetFormerRole(dm.GuildID, "")
		if err != nil {
			respond(fmt.Sprintf("Failed to clear former member role, %s", err))
			return
		}

		respond("Former role cleared")
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
			respond(fmt.Sprintf("Failed to set former role, %s", err))
			return
		}

		respond(fmt.Sprintf("Former role set to %s (`%s`)", roleName, roleID))
		return
	}

	formerRoleID := config.FormerRole(dm.GuildID)
	for _, role := range roles {
		if role.ID == formerRoleID {
			resp := fmt.Sprintf("Former role: %s (`%s`)", role.Name, role.ID)
			respond(resp)
			return
		}
	}

	respond("No Former Member role is configured")
}

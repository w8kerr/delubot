package mux

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/sheetsync"
)

func (m *Mux) Verify(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)
	msg := respond("=Processing verification...")
	edit := GetEditor(ds, msg)
	edit("```Processing verification...```")

	// Get last 10 messages
	msgs, err := ds.ChannelMessages(dm.ChannelID, 100, "", "", "")
	if err != nil {
		edit(fmt.Sprintf("```Failed to get channel messages, %s```", err))
		return
	}

	success := false
	handle := ""
	userID := ""
	proof := ""
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		success, handle, userID, proof = parseVerifMsg(msg)
		if success {
			break
		}
	}

	if !success {
		edit("```Could not verify, no attachments found```")
		return
	}

	planStr := strings.TrimSpace(strings.TrimPrefix(ctx.Content, "v"))
	plan, _ := strconv.Atoi(planStr)
	if plan == 0 {
		plan = 500
	}

	edit("```Granting roles...```")
	if plan >= 500 {
		alphaRole := config.AlphaRole(dm.GuildID)
		if alphaRole == "" {
			edit("```Could not verify, no Alpha role is configured```")
			return
		}
		err := ds.GuildMemberRoleAdd(dm.GuildID, userID, alphaRole)
		if err != nil {
			edit("```Could not verify, error adding Alpha role, " + err.Error() + "```")
			return
		}
	}
	if plan >= 1500 {
		specialRole := config.SpecialRole(dm.GuildID)
		if specialRole == "" {
			edit("```Could not verify, no Special role is configured```")
			return
		}
		err = ds.GuildMemberRoleAdd(dm.GuildID, userID, specialRole)
		if err != nil {
			edit("```Could not verify, error adding Special role, " + err.Error() + "```")
			return
		}
	}
	if plan >= 10000 {
		whaleRole := config.WhaleRole(dm.GuildID)
		if whaleRole == "" {
			edit("```Could not verify, no Whale role is configured```")
			return
		}
		err = ds.GuildMemberRoleAdd(dm.GuildID, userID, whaleRole)
		if err != nil {
			edit("```Could not verify, error adding Whale role, " + err.Error() + "```")
			return
		}
	}

	edit("```Updating Google Sheet...```")
	sheetID := config.SyncSheet(dm.GuildID)
	if sheetID == "" {
		edit("```Could not verify, no Google Sheet is configured, " + err.Error() + "```")
		return
	}
	sheetSvc, err := sheetsync.GetService()
	if err != nil {
		edit("```Could not verify, failed to connect to Google Sheet, " + err.Error() + "```")
		return
	}
	err = sheetsync.AddManualVerification(sheetSvc, sheetID, handle, userID, proof, plan, dm.Author.Username)
	if err != nil {
		edit("```Could not verify, error updating Google Sheet, " + err.Error() + "```")
		return
	}

	edit("```Recording log...```")
	logChanID := config.LogChannel(dm.GuildID)
	if logChanID != "" {
		logResp := "Handle:   " + handle
		logResp += "\nID:            " + userID
		logResp += "\nURL:         " + proof
		logResp += fmt.Sprintf("\nPlan:         %d", plan)
		logResp += "\nVerified:  " + dm.Author.Username
		_, err := ds.ChannelMessageSend(logChanID, logResp)
		if err != nil {
			log.Println("Failed to send log channel msg,", err)
		}
	}

	resp := "```Verification recorded"
	if plan >= 500 {
		resp += "\nAlpha role granted to " + handle
	}
	if plan >= 1500 {
		resp += "\nSpecial role granted to " + handle
	}
	if plan >= 10000 {
		resp += "\nWhale role granted to " + handle
	}
	resp += "\n\nYou may close the channel now```"

	edit(resp)
}

var footerRE = regexp.MustCompile(`^(.+) \| (\d{18})$`)

func parseVerifMsg(msg *discordgo.Message) (bool, string, string, string) {
	if len(msg.Attachments) == 0 {
		return false, "", "", ""
	}
	if len(msg.Embeds) == 0 {
		return false, "", "", ""
	}
	if msg.Embeds[0].Footer == nil {
		return false, "", "", ""
	}

	footer := msg.Embeds[0].Footer.Text
	footerMatch := footerRE.FindSubmatch([]byte(footer))
	if footerMatch == nil {
		return false, "", "", ""
	}
	handle := string(footerMatch[1])
	userID := string(footerMatch[2])

	if handle == "" || userID == "" {
		return false, "", "", ""
	}

	attachments := []string{}
	for _, a := range msg.Attachments {
		attachments = append(attachments, a.URL)
	}

	if len(attachments) == 0 {
		return false, "", "", ""
	}

	proof := strings.Join(attachments, ", ")

	return true, handle, userID, proof
}

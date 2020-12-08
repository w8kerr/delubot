package mux

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/sheetsync"
	"github.com/w8kerr/delubot/utils"
)

func (m *Mux) Verify(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)
	msg := respond("=🔺Processing verification...")
	edit := GetEditor(ds, msg)
	edit("```🔺Processing verification...```")

	// Get last 10 messages
	msgs, err := ds.ChannelMessages(dm.ChannelID, 100, "", "", "")
	if err != nil {
		edit(fmt.Sprintf("```Failed to get channel messages, %s```", err))
		return
	}

	handle := ""
	userID := ""
	proofs := []string{}
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		h, uID, attachments, err := parseVerifMsg(msg)
		if h != "" && uID != "" {
			handle = h
			userID = uID
		}
		if err == nil {
			proofs = append(proofs, attachments...)
		}
	}

	planStr := strings.TrimSpace(strings.TrimPrefix(ctx.Content, "v"))
	parts := strings.Split(planStr, " ")
	plan, _ := strconv.Atoi(parts[0])
	if plan == 0 {
		plan = 500
	}

	if len(parts) > 1 {
		proofs = parts[1:]
	}

	if len(proofs) == 0 {
		edit("```Could not verify, no attachments found```")
		return
	}

	proof := strings.Join(proofs, " | ")

	formerRoleError := false

	edit("```🔺Granting roles...```")
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
		formerRole := config.FormerRole(dm.GuildID)
		if formerRole != "" {
			err = ds.GuildMemberRoleRemove(dm.GuildID, userID, formerRole)
			if err != nil {
				formerRoleError = true
			}
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

	edit("```🔺Updating Google Sheet...```")
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

	edit("```🔺Recording log...```")
	logChanID := config.LogChannel(dm.GuildID)
	if logChanID != "" {
		logResp := "Handle:   " + handle
		logResp += "\nID:            " + userID
		logResp += "\nProof:      " + proof
		logResp += fmt.Sprintf("\nPlan:         %d", plan)
		logResp += "\nVerified:  " + dm.Author.Username
		_, err := ds.ChannelMessageSend(logChanID, logResp)
		if err != nil {
			log.Println("Failed to send log channel msg,", err)
		}
	}

	resp := "```🔺Verification recorded"
	if plan >= 500 {
		resp += "\nAlpha role granted to   " + handle
	}
	if plan >= 1500 {
		resp += "\nSpecial role granted to " + handle
	}
	if plan >= 10000 {
		resp += "\nWhale role granted to   " + handle
	}
	if formerRoleError {
		resp += "\n(Failed to remove Former Member role, you'll have to do that yourself)"
	}
	resp += "\n\nYou may close the channel now```"

	edit(resp)
}

func (m *Mux) VDebug(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)
	respond("=🔺Debugging verification (check internal logs)...")
	fmt.Println("```🔺Processing verification...```")

	// Get last 10 messages
	msgs, err := ds.ChannelMessages(dm.ChannelID, 100, "", "", "")
	if err != nil {
		fmt.Println(fmt.Sprintf("```Failed to get channel messages, %s```", err))
		return
	}

	resp := "=🔺Messages in this channel:"

	proofs := []string{}
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		msgLines := strings.Split(msg.Content, "\n")
		resp += fmt.Sprintf("\n> Message %d (%s)", i, msgLines[0])
		_, _, attachments, err := parseVerifMsg(msg)
		if err == nil {
			proofs = append(proofs, attachments...)

			line := "Extracted proofs: " + strings.Join(attachments, " | ")
			resp += "\n" + line
			fmt.Println(line)
		} else {
			line := err.Error()
			resp += "\n" + line
			fmt.Println(line)
		}

		utils.PrintJSON(msg)
	}

	if len(proofs) > 0 {
		resp += "\n\nFinal extracted proofs: " + strings.Join(proofs, " | ")
	} else {
		resp += "\n\nNo proofs extracted"
	}

	respond(resp)
}

var footerRE = regexp.MustCompile(`^(.+) \| (\d{16,18})$`)

func parseVerifMsg(msg *discordgo.Message) (string, string, []string, error) {
	if len(msg.Embeds) == 0 {
		return "", "", []string{}, errors.New("No embeds")
	}
	if msg.Embeds[0].Footer == nil {
		return "", "", []string{}, errors.New("No embed footer")
	}
	if msg.Embeds[0].Title != "Message Received" {
		return "", "", []string{}, errors.New("Embed title was not 'Message Received'")
	}

	footer := msg.Embeds[0].Footer.Text
	footerMatch := footerRE.FindSubmatch([]byte(footer))
	if footerMatch == nil {
		return "", "", []string{}, errors.New("Footer did not match expected pattern")
	}
	handle := string(footerMatch[1])
	userID := string(footerMatch[2])

	if handle == "" || userID == "" {
		return "", "", []string{}, fmt.Errorf("Either handle (%s) or userID (%s) was nil", handle, userID)
	}

	if len(msg.Attachments) == 0 {
		return handle, userID, []string{}, errors.New("No attachments")
	}

	attachments := []string{}
	for _, a := range msg.Attachments {
		attachments = append(attachments, a.URL)
	}

	if len(attachments) == 0 {
		return handle, userID, []string{}, errors.New("No attachments found in the end")
	}

	return handle, userID, attachments, nil
}

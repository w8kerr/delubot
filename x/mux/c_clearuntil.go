package mux

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/utils"
)

func (m *Mux) ClearUntil(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	channel, err := ds.Channel(dm.ChannelID)
	if err != nil {
		respond(fmt.Sprintf("🔺Failed to get channel: %s", err))
		return
	}

	repliedID := ""
	if dm.MessageReference != nil {
		repliedID = dm.MessageReference.MessageID
	} else {
		respond("🔺You must Reply to the earliest message you want to clear\n(please turn off the notification to avoid confusing that person)")
		return
	}

	lastMessageID := channel.LastMessageID
	finished := false
	messageIDs := []string{}
	text := ""

	for !finished {
		messages, err := ds.ChannelMessages(dm.ChannelID, 100, lastMessageID, "", "")
		if err != nil {
			respond(fmt.Sprintf("🔺Failed to get messages from channel: %s", err))
			return
		}

		if len(messages) == 0 {
			respond("🔺Failed to find the replied-to message")
			return
		}

		for _, message := range messages {
			AddMessageToText(message, &text)
			messageIDs = append(messageIDs, message.ID)

			if message.ID == repliedID {
				finished = true
				break
			}
		}

		if messages[len(messages)-1].ID == lastMessageID {
			respond("🔺Unexpectedly received a repeat batch of message")
			return
		}

		lastMessageID = messages[len(messages)-1].ID
	}

	dmChannel, err := ds.UserChannelCreate(dm.Author.ID)
	if err != nil {
		respond("Failed to DM cleared messages")
		utils.OutputTextToFile(ds, dm.ChannelID, "cleared_messages.txt", text)
	} else {
		utils.OutputTextToFile(ds, dmChannel.ID, "cleared_messages.txt", text)
	}
	m.DoClear(ds, dm.ChannelID, messageIDs)
	respond(fmt.Sprintf("🔺Cleared %d messages", len(messageIDs)))
}

func (m *Mux) DoClear(ds *discordgo.Session, channelID string, deleteMessageIDs []string) {
	err := utils.BulkDeleteMessages(ds, channelID, deleteMessageIDs)
	if err != nil {
		ds.ChannelMessageSend(channelID, fmt.Sprintf("Failed to delete messages: %s", err))
		return
	}
}

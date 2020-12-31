package mux

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/utils"
)

func (m *Mux) ExtractMessages(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	ctx.Content = strings.TrimPrefix(ctx.Content, "extractmessages")
	ctx.Content = strings.TrimSpace(ctx.Content)

	if ctx.Content == "" {
		respond("🔺Usage: -db extractmessages <text that should match the beginning and end of extracted section>")
		return
	}

	channel, err := ds.Channel(dm.ChannelID)
	if err != nil {
		respond(fmt.Sprintf("🔺Failed to get channel: %s", err))
		return
	}

	extracting := false
	lastMessageID := channel.LastMessageID
	finished := false
	firstComment := ""
	lastComment := ""
	messageIDs := []string{}
	text := ""

	for !finished {
		messages, err := ds.ChannelMessages(dm.ChannelID, 100, lastMessageID, "", "")
		if err != nil {
			respond(fmt.Sprintf("🔺Failed to get messages from channel: %s", err))
			return
		}

		if len(messages) == 0 {
			respond(fmt.Sprintf("🔺Failed to find two instances of the search string ❝ %s ❞", ctx.Content))
			return
		}

		for _, message := range messages {
			if message.ID == dm.ID {
				break
			}
			if strings.Contains(message.Content, ctx.Content) {
				if !extracting {
					// Start extracting
					extracting = true
					lastComment = message.Content
				} else {
					// Finish extracting
					finished = true
					firstComment = message.Content
				}
			}

			if extracting {
				AddMessageToText(message, &text)
				messageIDs = append(messageIDs, message.ID)
			}

			if finished {
				break
			}
		}

		if messages[len(messages)-1].ID == lastMessageID {
			respond("🔺Unexpectedly received a repeat batch of message")
			return
		}

		lastMessageID = messages[len(messages)-1].ID
	}

	utils.OutputTextToFile(ds, dm.ChannelID, "extracted_messages.txt", text)
	msg := respond(fmt.Sprintf("🔺Delete %d messages from:\n❝ %s ❞\nto\n❝ %s ❞?", len(messageIDs), firstComment, lastComment))
	err = ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\U0001F1FE")
	if err != nil {
		fmt.Printf("Failed to add \U0001F1FE reaction, %s\n", err)
	}
	err = ds.MessageReactionAdd(msg.ChannelID, msg.ID, "\U0001F1F3")
	if err != nil {
		fmt.Printf("Failed to add \U0001F1F3 reaction, %s\n", err)
	}
	extraction := config.Extraction{
		ChannelID:         dm.ChannelID,
		UserMessageID:     dm.ID,
		BotMessageID:      msg.ID,
		ExtractMessageIDs: messageIDs,
	}
	config.Extractions[msg.ID] = extraction
}

func AddMessageToText(message *discordgo.Message, text *string) {
	t, _ := message.Timestamp.Parse()
	line := fmt.Sprintf("[%s] %s: %s\n", t.In(config.Loc).Format("06/1/2 15:04:05"), message.Author.Username, message.Content)
	*text = line + *text
}

func (m *Mux) CancelExtraction(ds *discordgo.Session, e config.Extraction) {
	ds.ChannelMessageDelete(e.ChannelID, e.UserMessageID)
	ds.ChannelMessageDelete(e.ChannelID, e.BotMessageID)
}

func (m *Mux) DoExtraction(ds *discordgo.Session, e config.Extraction) {
	err := utils.BulkDeleteMessages(ds, e.ChannelID, e.ExtractMessageIDs)
	if err != nil {
		ds.ChannelMessageSend(e.ChannelID, fmt.Sprintf("Failed to delete messages: %s", err))
		return
	}

	ds.ChannelMessageDelete(e.ChannelID, e.UserMessageID)
	ds.ChannelMessageDelete(e.ChannelID, e.BotMessageID)
}

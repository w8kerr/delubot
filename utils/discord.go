package utils

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func GetAllMembers(ds *discordgo.Session, guildID string) ([]*discordgo.Member, error) {
	limit := 1000
	after := ""
	lastMember := false
	list := []*discordgo.Member{}

	for !lastMember {
		memberChunk, err := ds.GuildMembers(guildID, after, limit)
		if err != nil {
			log.Printf("Error getting guild members, %s", err)
			return list, err
		}

		if len(memberChunk) == 0 {
			lastMember = true
			break
		}

		after = memberChunk[len(memberChunk)-1].User.ID

		if len(memberChunk) < limit {
			lastMember = true
		} else if len(list) > 0 && len(memberChunk) > 0 && memberChunk[len(memberChunk)-1].User.ID == list[len(list)-1].User.ID {
			lastMember = true
		}

		list = append(list, memberChunk...)
	}

	return list, nil
}

type ChannelLogger struct {
	Session   *discordgo.Session
	ChannelID string
}

func (cl ChannelLogger) Write(p []byte) (n int, err error) {
	_, err = cl.Session.ChannelMessageSend(cl.ChannelID, string(p))
	return len(p), err
}

func GetChannelLogger(ds *discordgo.Session, channelID string) *log.Logger {
	cl := ChannelLogger{
		Session:   ds,
		ChannelID: channelID,
	}

	return log.New(cl, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func OutputTextToFile(ds *discordgo.Session, channelID, filename, text string) {
	reader := strings.NewReader(text)
	ds.ChannelFileSend(channelID, filename, reader)
}

func BulkDeleteMessages(ds *discordgo.Session, channelID string, messageIDs []string) error {
	for len(messageIDs) > 100 {
		batch := messageIDs[0:99]
		err := ds.ChannelMessagesBulkDelete(channelID, batch)
		if err != nil {
			return err
		}
		messageIDs = messageIDs[100:]
	}

	err := ds.ChannelMessagesBulkDelete(channelID, messageIDs)
	if err != nil {
		return err
	}
	return nil
}

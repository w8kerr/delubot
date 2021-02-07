package mux

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
)

// IsModerator check if a user is a moderator
func IsModerator(ds *discordgo.Session, dm *discordgo.MessageCreate) bool {
	member, err := ds.GuildMember(dm.GuildID, dm.Author.ID)
	if err != nil {
		log.Printf("error getting user's member, %s", err)
		return false
	}

	guildMods, ok := config.ModeratorRoles[dm.GuildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", dm.GuildID)
		return false
	}

	for _, modRole := range guildMods {
		for _, memberRole := range member.Roles {
			if modRole == memberRole {
				return true
			}
		}
	}

	return false
}

func GetResponder(ds *discordgo.Session, dm *discordgo.Message) func(msg string) *discordgo.Message {
	return func(msg string) *discordgo.Message {
		msgParts := []string{}

		runes := []rune(msg)

		for len(runes) > 2000 {
			msgParts = append(msgParts, string(runes[0:2000]))
			runes = runes[2000:]
		}
		msgParts = append(msgParts, string(runes))

		var ret *discordgo.Message
		var err error
		for i, part := range msgParts {
			ret, err = ds.ChannelMessageSend(dm.ChannelID, part)
			if err != nil {
				fmt.Println(i, len(part), err)
				fmt.Println(part)
			}
		}

		return ret
	}
}

func GetEditor(ds *discordgo.Session, dm *discordgo.Message) func(msg string) *discordgo.Message {
	return func(msg string) *discordgo.Message {
		ret, err := ds.ChannelMessageEdit(dm.ChannelID, dm.ID, msg)
		if err != nil {
			fmt.Println("Error editing message:", err)
		}
		return ret
	}
}

func PadString(str string, length int) string {
	if len(str) > length {
		return str
	}

	padding := length - len(str)
	str = str + strings.Repeat(" ", padding)

	return str
}

func NumLength(in int) int {
	abs := in
	if in == 0 {
		return 1
	}

	if in < 0 {
		abs = -1 * in
	}

	length := math.Ceil(math.Log10(float64(abs + 1)))
	if in < 0 {
		length++
	}

	return int(length)
}

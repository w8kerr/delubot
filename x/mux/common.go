package mux

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/models"
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

// IsStaff check if a user is a staff member
func IsStaff(ds *discordgo.Session, dm *discordgo.MessageCreate) bool {
	member, err := ds.GuildMember(dm.GuildID, dm.Author.ID)
	if err != nil {
		log.Printf("error getting user's member, %s", err)
		return false
	}

	guildMods, ok := config.StaffRoles[dm.GuildID]
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

// IsStaffReaction check if a user is a staff member (on a reaction rather than a message)
func IsStaff(ds *discordgo.Session, ra *discordgo.MessageReactionAdd) bool {
	member, err := ds.GuildMember(ra.GuildID, ra.UserID)
	if err != nil {
		log.Printf("error getting user's member, %s", err)
		return false
	}

	guildMods, ok := config.StaffRoles[ra.GuildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", ra.GuildID)
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

// HasAccess check if a user has the specified access level
func HasAccess(ds *discordgo.Session, dm *discordgo.MessageCreate, access int) bool {
	switch access {
	case models.AL_EVERYONE:
		return true
	case models.AL_STAFF:
		return IsStaff(ds, dm)
	case models.AL_MOD:
		return IsModerator(ds, dm)
	case models.AL_DEV:
		return dm.Author.ID == config.CreatorID
	default:
		log.Println("Command with unknown access level!")
		return false
	}
}

func GetAccessSymbol(access int) string {
	switch access {
	case models.AL_EVERYONE:
		return "α"
	case models.AL_STAFF:
		return "Ψ"
	case models.AL_MOD:
		return "θ"
	case models.AL_DEV:
		return "M"
	default:
		return "?"
	}
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

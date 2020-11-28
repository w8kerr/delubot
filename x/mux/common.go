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

		fmt.Println("MEMBERCHUNK", len(memberChunk))

		if len(memberChunk) < limit {
			lastMember = true
		} else if len(list) > 0 && len(memberChunk) > 0 && memberChunk[len(memberChunk)-1].User.ID == list[len(list)-1].User.ID {
			lastMember = true
		}

		list = append(list, memberChunk...)
	}

	return list, nil
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

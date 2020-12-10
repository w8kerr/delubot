package utils

import (
	"log"

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

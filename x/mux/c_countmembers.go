package mux

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
)

func (m *Mux) CountMembers(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := func(msg string) {
		_, err := ds.ChannelMessageSend(dm.ChannelID, msg)
		if err != nil {
			fmt.Println(err)
		}
	}

	// guild, err := ds.Guild(dm.GuildID)
	// if err != nil {
	// 	respond(err.Error())
	// 	return
	// }

	members, err := GetAllMembers(ds, dm.GuildID)
	if err != nil {
		respond("Error: " + err.Error())
		return
	}

	roleMap := make(map[string]int)

	for _, member := range members {
		for _, role := range member.Roles {
			if _, ok := roleMap[role]; !ok {
				roleMap[role] = 0
			}

			roleMap[role] = roleMap[role] + 1
		}
	}

	roles, err := ds.GuildRoles(dm.GuildID)
	if err != nil {
		respond("Error: " + err.Error())
		return
	}

	resp := "Count members!\n```"
	resp += fmt.Sprintf("All - %d\n\n", len(members))
	maxLength := 0
	for _, role := range roles {
		if count, ok := roleMap[role.ID]; ok {
			numLength := NumLength(count)
			if numLength > maxLength {
				maxLength = numLength
			}
		}
	}

	sort.SliceStable(roles, func(i, j int) bool {
		return roles[i].Position > roles[j].Position
	})

	for _, role := range roles {
		if role.Name == "@everyone" {
			continue
		}
		if count, ok := roleMap[role.ID]; ok {
			line := PadString(fmt.Sprintf("%d", count), maxLength) + fmt.Sprintf(" | %s\n", role.Name)
			resp += line
		} else {
			line := PadString("0", maxLength) + fmt.Sprintf(" | %s\n", role.Name)
			resp += line
		}
	}
	resp += "```"

	respond(resp)
	if err != nil {
		fmt.Println(err)
	}
}

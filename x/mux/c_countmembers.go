package mux

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/utils"
)

func (m *Mux) CountMembers(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	prerespond := GetResponder(ds, dm)
	msg := prerespond("🔺Looking up member information...")
	respond := GetEditor(ds, msg)

	// guild, err := ds.Guild(dm.GuildID)
	// if err != nil {
	// 	respond(err.Error())
	// 	return
	// }

	members, err := utils.GetAllMembers(ds, dm.GuildID)
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
		line := ""
		if count, ok := roleMap[role.ID]; ok {
			line = PadString(fmt.Sprintf("%d", count), maxLength) + fmt.Sprintf(" | %s\n", role.Name)
		} else {
			line = PadString("0", maxLength) + fmt.Sprintf(" | %s\n", role.Name)
		}

		if len(resp)+len(line) > 1997 {
			respond(resp + "```")
			respond = prerespond
			resp = "```"
		}
		resp += line
	}
	resp += "```"

	respond(resp)
	if err != nil {
		fmt.Println(err)
	}
}

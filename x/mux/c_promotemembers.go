package mux

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/sheetsync"
	"github.com/w8kerr/delubot/utils"
)

func (m *Mux) PromoteMembers(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	prerespond := GetResponder(ds, dm)
	msg := prerespond("🔺Promoting expired members...")
	respond := GetEditor(ds, msg)

	guildID := "755437328515989564"
	alphaRole := config.AlphaRole(guildID)
	formerRole := config.FormerRole(guildID)

	members, err := utils.GetAllMembers(ds, guildID)
	if err != nil {
		respond(fmt.Sprintf("🔺Failed to get guild members: %s", err))
		return
	}

	num := 0
	updateTime := time.Now()
	for _, member := range members {
		if sheetsync.HasRole(member, formerRole) {
			num++
			if !sheetsync.HasRole(member, alphaRole) {
				err := ds.GuildMemberRoleAdd(guildID, member.User.ID, alphaRole)
				if err != nil {
					log.Printf("Failed to add alpha role %s", err)
				}
			}
			err = ds.GuildMemberRoleRemove(guildID, member.User.ID, formerRole)
			if err != nil {
				log.Printf("Failed to remove expired role %s", err)
			}
			if updateTime.Add(5 * time.Second).After(time.Now()) {
				updateTime = time.Now()
				respond(fmt.Sprintf("🔺Processed %d expired members...", num))
			}
		}
	}

	respond(fmt.Sprintf("🔺Promoted all expired members (%d) to Alpha!", num))
}

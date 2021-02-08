package mux

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/utils"
)

func (m *Mux) Mods(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
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

	guildMods, ok := config.ModeratorRoles[dm.GuildID]
	if !ok {
		respond("Could not find guild roles: " + dm.GuildID)
		return
	}

	mods := []*discordgo.Member{}

	isMod := func(member *discordgo.Member) bool {
		for _, modRole := range guildMods {
			for _, memberRole := range member.Roles {
				if modRole == memberRole {
					return true
				}
			}
		}
		return false
	}

	for _, member := range members {
		if isMod(member) {
			mods = append(mods, member)
		}
	}

	sort.SliceStable(mods, func(i, j int) bool {
		return mods[i].User.Username > mods[j].User.Username
	})

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)
	adminCol := db.C("admins")

	resp := "🔺Current moderators!\n```"
	for _, mod := range mods {
		name := mod.User.Username + "#" + mod.User.Discriminator
		resp += mod.User.ID + " | " + name
		_, err := adminCol.Upsert(bson.M{"discord_id": mod.User.ID}, bson.M{"$set": bson.M{"discord_name": name}})
		if err != nil {
			fmt.Println("Failed to set admin record:", err)
			resp += " (" + err.Error() + ")"
		}
		resp += "\n"
	}

	respond(resp + "```")
	if err != nil {
		fmt.Println(err)
	}
}

package mux

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
)

var Triangle = "\U0001F53A"

func (m *Mux) Proposal(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)
	msg := respond("```Sign off sheet\n--------------\nReact to sign off on this proposal```")
	config.Proposals[msg.ID] = msg.ChannelID
	ds.MessageReactionAdd(msg.ChannelID, msg.ID, Triangle)
}

func (m *Mux) UpdateProposal(ds *discordgo.Session, guildID, channelID, messageID string) {
	users, err := ds.MessageReactions(channelID, messageID, Triangle, 100)
	if err != nil {
		fmt.Println("ERROR GETTING REACTIONS ON PROPOSAL", channelID, messageID, err.Error())
	}

	msg := "```Sign off sheet\n--------------\n"
	if len(users) == 0 {
		msg += "React to sign off on this proposal"
		ds.MessageReactionAdd(channelID, messageID, Triangle)
	} else {
		for _, user := range users {
			fmt.Println("Process user", user.Username)
			// Don't count DeluBot as a signer, and remove its reaction when someone else's is there
			if user.ID == ds.State.User.ID {
				ds.MessageReactionRemove(channelID, messageID, Triangle, ds.State.User.ID)
				continue
			}
			member, err := ds.GuildMember(guildID, user.ID)
			if err != nil {
				fmt.Println("ERROR GETTING MEMBER OF REACTION", channelID, messageID, err.Error())
				continue
			}
			if member.Nick != "" {
				msg += member.Nick + "\n"
			} else {
				msg += user.Username + "\n"
			}
		}
	}
	msg += "```"

	ds.ChannelMessageEdit(channelID, messageID, msg)
}

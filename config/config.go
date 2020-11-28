/*
 * Realtime config from the DB
 */

package config

import (
	"context"
	"log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/w8kerr/delubot/models"
	"github.com/w8kerr/delubot/mongo"
)

// Get Load the config object
func Get(c context.Context) models.Config {
	session := mongo.MDB.Clone()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)

	configCol := db.C("config")
	config := models.Config{}
	configCol.Find(bson.M{}).One(&config)

	return config
}

var ModeratorRoles = map[string][]string{
	"782092598290546719": { // DTS
		"782109193435611138", // DTS Moderator
	},
	"755437328515989564": { // DFS
		"755623281238867980", // Founder - Ω
		"775181732864852048", // Server Admin Team - Φ
		"778288741504516096", // Head Mod
		"775667196969877504", // [WIP] Engineer
		"755788358994755664", // Moderators - θ
		"770268062871715850", // JP Mods/モデレータ - θ
		"770266890882383912", // CN Mods/管理員 - θ
		"770040458666836049", // Reddit Mods - θ
	},
}

type RoleConfig struct {
	Alpha string
	Whale string
}

var grantRoles = map[string]RoleConfig{
	"782092598290546719": { // DTS
		Alpha: "782115624537030689",
		Whale: "782115760994517003",
	},
	"755437328515989564": { // DFS
		Alpha: "760705266953355295",
		Whale: "761570574794489886",
	},
}

// AlphaRole get the designated alpha role for the given guild
func AlphaRole(guildID string) string {
	guildRoles, ok := grantRoles[guildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", guildID)
		return ""
	}

	return guildRoles.Alpha
}

// WhaleRole get the designated alpha role for the given guild
func WhaleRole(guildID string) string {
	guildRoles, ok := grantRoles[guildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", guildID)
		return ""
	}

	return guildRoles.Whale
}

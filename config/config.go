/*
 * Realtime config from the DB
 */

package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/w8kerr/delubot/models"
	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/utils"
)

var ModeratorRoles = map[string][]string{
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
	Alpha string `json:"alpha" bson:"alpha"`
	Whale string `json:"whale" bson:"whale"`
}

var GrantRoles = map[string]RoleConfig{
	"755437328515989564": { // DFS
		Alpha: "760705266953355295",
		Whale: "761570574794489886",
	},
}

var SyncSheets = map[string]string{}

var SyncEnabled = map[string]bool{}

var TimeFormat string

var GoogleCredentials bson.M

var Loc *time.Location

type BotConfig struct {
	ModeratorRoles    map[string][]string   `json:"moderator_roles" bson:"moderator_roles"`
	GrantRoles        map[string]RoleConfig `json:"grant_roles" bson:"grant_roles"`
	SyncSheets        map[string]string     `json:"sync_sheets" bson:"sync_sheets"`
	SyncEnabled       map[string]bool       `json:"sync_enabled" bson:"sync_enabled"`
	TimeFormat        string                `json:"time_format" bson:"time_format"`
	GoogleCredentials bson.M                `json:"-" bson:"google_credentials"`
}

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

func LoadConfig() error {

	defaultConfig := BotConfig{
		ModeratorRoles: ModeratorRoles,
		GrantRoles:     GrantRoles,
		SyncSheets:     SyncSheets,
		TimeFormat:     TimeFormat,
	}

	fmt.Println("DEFAULT CONFIG")
	utils.PrintJSON(defaultConfig)

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)

	configCol := db.C("config")
	config := BotConfig{}
	err := configCol.Find(bson.M{}).One(&config)
	if err != nil {
		log.Printf("Failed to load config, %s", err)
		return err
	}

	fmt.Println("CONFIG FROM DB")
	utils.PrintJSON(config)

	ModeratorRoles = config.ModeratorRoles
	GrantRoles = config.GrantRoles
	SyncSheets = config.SyncSheets
	SyncEnabled = config.SyncEnabled
	TimeFormat = config.TimeFormat
	GoogleCredentials = config.GoogleCredentials

	Loc, _ = time.LoadLocation("Asia/Tokyo")

	return nil
}

func UpdateConfig(update bson.M) error {
	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)

	fmt.Println("SET CONFIG")
	utils.PrintJSON(update)

	configCol := db.C("config")
	err := configCol.Update(bson.M{}, bson.M{"$set": update})
	return err
}

// SyncSheet get the designated sync sheet for the given guild
func SyncSheet(guildID string) string {
	syncSheet, ok := SyncSheets[guildID]
	if !ok {
		log.Printf("Could not find sync sheet, %s", guildID)
		return ""
	}

	return syncSheet
}

// SyncIsEnabled Return whether syncing is enabled for the given guild
func SyncIsEnabled(guildID string) bool {
	syncEnabled, ok := SyncEnabled[guildID]
	if !ok {
		return false
	}

	return syncEnabled
}

// AlphaRole get the designated alpha role for the given guild
func AlphaRole(guildID string) string {
	guildRoles, ok := GrantRoles[guildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", guildID)
		return ""
	}

	return guildRoles.Alpha
}

// WhaleRole get the designated alpha role for the given guild
func WhaleRole(guildID string) string {
	guildRoles, ok := GrantRoles[guildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", guildID)
		return ""
	}

	return guildRoles.Whale
}

func SetAlphaRole(guildID, roleID string) error {
	key := fmt.Sprintf("grant_roles.%s.alpha", guildID)
	update := bson.M{
		key: roleID,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	if v, ok := GrantRoles[guildID]; ok {
		v.Alpha = roleID
		GrantRoles[guildID] = v
	} else {
		GrantRoles[guildID] = RoleConfig{
			Alpha: roleID,
		}
	}

	return nil
}

func SetWhaleRole(guildID, roleID string) error {
	key := fmt.Sprintf("grant_roles.%s.whale", guildID)
	update := bson.M{
		key: roleID,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	if v, ok := GrantRoles[guildID]; ok {
		v.Whale = roleID
		GrantRoles[guildID] = v
	} else {
		GrantRoles[guildID] = RoleConfig{
			Whale: roleID,
		}
	}

	return nil
}

func SetSyncSheet(guildID, sheetID string) error {
	key := fmt.Sprintf("sync_sheets.%s", guildID)
	update := bson.M{
		key: sheetID,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	SyncSheets[guildID] = sheetID

	return nil
}

func SetSyncEnabled(guildID string, enabled bool) error {
	key := fmt.Sprintf("sync_enabled.%s", guildID)
	update := bson.M{
		key: enabled,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	SyncEnabled[guildID] = enabled

	return nil
}

func ParseTime(raw string) time.Time {
	parsed, err := time.ParseInLocation(TimeFormat, raw, Loc)
	if err != nil {
		log.Printf("Invalid end time, %s", raw)
		return time.Time{}
	}

	return parsed
}

func PrintTime(t time.Time) string {
	return t.Format(TimeFormat)
}

func Now() time.Time {
	return time.Now().In(Loc)
}

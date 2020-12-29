/*
 * Realtime config from the DB
 */

package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
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
	Alpha   string `json:"alpha" bson:"alpha"`
	Special string `json:"special" bson:"special"`
	Whale   string `json:"whale" bson:"whale"`
	Former  string `json:"former" bson:"former"`
}

type TweetSyncConfig struct {
	Handle    string `json:"handle" bson:"handle"`
	ChannelID string `json:"channel_id" bson:"channel_id"`
	SinceID   int64  `json:"since_id" bson:"since_id"`
}

type TweetUpdate struct {
	ChannelID      string `json:"channel_id" bson:"channel_id"`
	UserMessageID  string `json:"user_message_id" bson:"user_message_id"`
	BotMessageID   string `json:"bot_message_id" bson:"bot_message_id"`
	TweetMessageID string `json:"tweet_message_id" bson:"tweet_message_id"`
	Translation    string `json:"translation" bson:"translation"`
	Translator     string `json:"translator" bson:"translator"`
}

var GrantRoles = map[string]RoleConfig{
	"755437328515989564": { // DFS
		Alpha:   "760705266953355295",
		Special: "783112023570513970",
		Whale:   "761570574794489886",
		Former:  "782479385185615953",
	},
}

var SyncSheets = map[string]string{}

var RoleGrantEnabled = map[string]bool{}
var RoleRemoveEnabled = map[string]bool{}

var ModmailCategories = map[string]string{
	"755437328515989564": "779849308525690900",
}

var LogChannels = map[string]string{
	"755437328515989564": "772322798546321428",
}

var ErrorChannel = "793361959046217778"

var TimeFormat string

var GoogleCredentials bson.M

var EightBallEnabled bool

var Loc *time.Location

var Proposals = make(map[string]string)

var CreatorID = "204752740503650304"

var TweetSyncChannels = []TweetSyncConfig{}

var TweetUpdates = make(map[string]TweetUpdate)

type BotConfig struct {
	ModeratorRoles    map[string][]string   `json:"moderator_roles" bson:"moderator_roles"`
	GrantRoles        map[string]RoleConfig `json:"grant_roles" bson:"grant_roles"`
	SyncSheets        map[string]string     `json:"sync_sheets" bson:"sync_sheets"`
	RoleGrantEnabled  map[string]bool       `json:"role_grant_enabled" bson:"role_grant_enabled"`
	RoleRemoveEnabled map[string]bool       `json:"role_remove_enabled" bson:"role_remove_enabled"`
	TimeFormat        string                `json:"time_format" bson:"time_format"`
	GoogleCredentials bson.M                `json:"-" bson:"google_credentials"`
	EightBallEnabled  bool                  `json:"eight_ball_enabled" bson:"eight_ball_enabled"`
	TweetSyncChannels []TweetSyncConfig     `json:"tweet_sync_channels" bson:"tweet_sync_channels"`
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

	ModeratorRoles = config.ModeratorRoles
	GrantRoles = config.GrantRoles
	SyncSheets = config.SyncSheets
	RoleGrantEnabled = config.RoleGrantEnabled
	RoleRemoveEnabled = config.RoleRemoveEnabled
	TimeFormat = config.TimeFormat
	GoogleCredentials = config.GoogleCredentials
	EightBallEnabled = config.EightBallEnabled
	TweetSyncChannels = config.TweetSyncChannels

	if GrantRoles == nil {
		GrantRoles = make(map[string]RoleConfig)
	}
	if SyncSheets == nil {
		SyncSheets = make(map[string]string)
	}
	if RoleGrantEnabled == nil {
		RoleGrantEnabled = make(map[string]bool)
	}
	if RoleRemoveEnabled == nil {
		RoleRemoveEnabled = make(map[string]bool)
	}

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

// RoleGrantIsEnabled Return whether syncing is enabled for the given guild
func RoleGrantIsEnabled(guildID string) bool {
	grantEnabled, ok := RoleGrantEnabled[guildID]
	if !ok {
		return false
	}

	return grantEnabled
}

// RoleRemoveIsEnabled Return whether syncing is enabled for the given guild
func RoleRemoveIsEnabled(guildID string) bool {
	removeEnabled, ok := RoleRemoveEnabled[guildID]
	if !ok {
		return false
	}

	return removeEnabled
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

// SpecialRole get the designated special role for the given guild
func SpecialRole(guildID string) string {
	guildRoles, ok := GrantRoles[guildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", guildID)
		return ""
	}

	return guildRoles.Special
}

// WhaleRole get the designated whale role for the given guild
func WhaleRole(guildID string) string {
	guildRoles, ok := GrantRoles[guildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", guildID)
		return ""
	}

	return guildRoles.Whale
}

// FormerRole get the designated former member role for the given guild
func FormerRole(guildID string) string {
	guildRoles, ok := GrantRoles[guildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", guildID)
		return ""
	}

	return guildRoles.Former
}

// ModmailCategory get the designated modmail category ID for the given guild
func ModmailCategory(guildID string) string {
	catID, ok := ModmailCategories[guildID]
	if !ok {
		log.Printf("Could not find modmail category, %s", guildID)
		return ""
	}

	return catID
}

func IsModmailChannel(ds *discordgo.Session, guildID, channelID string) bool {
	catID, ok := ModmailCategories[guildID]
	if !ok {
		log.Printf("Could not find modmail category, %s", guildID)
		return false
	}

	ch, err := ds.Channel(channelID)
	if err != nil {
		log.Printf("Failed to get channel info, %s", err)
		return false
	}
	return ch.ParentID == catID
}

// LogChannel get the designated log channel for the given guild
func LogChannel(guildID string) string {
	chanID, ok := LogChannels[guildID]
	if !ok {
		log.Printf("Could not find modmail category, %s", guildID)
		return ""
	}

	return chanID
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

func SetSpecialRole(guildID, roleID string) error {
	key := fmt.Sprintf("grant_roles.%s.special", guildID)
	update := bson.M{
		key: roleID,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	if v, ok := GrantRoles[guildID]; ok {
		v.Special = roleID
		GrantRoles[guildID] = v
	} else {
		GrantRoles[guildID] = RoleConfig{
			Special: roleID,
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

func SetFormerRole(guildID, roleID string) error {
	key := fmt.Sprintf("grant_roles.%s.former", guildID)
	update := bson.M{
		key: roleID,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	if v, ok := GrantRoles[guildID]; ok {
		v.Former = roleID
		GrantRoles[guildID] = v
	} else {
		GrantRoles[guildID] = RoleConfig{
			Former: roleID,
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

func SetRoleGrantEnabled(guildID string, enabled bool) error {
	key := fmt.Sprintf("role_grant_enabled.%s", guildID)
	update := bson.M{
		key: enabled,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	RoleGrantEnabled[guildID] = enabled

	return nil
}

func SetRoleRemoveEnabled(guildID string, enabled bool) error {
	key := fmt.Sprintf("role_remove_enabled.%s", guildID)
	update := bson.M{
		key: enabled,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	RoleRemoveEnabled[guildID] = enabled

	return nil
}

func SetEightBallEnabled(enabled bool) error {
	update := bson.M{
		"eight_ball_enabled": enabled,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	EightBallEnabled = enabled
	return nil
}

func SetTweetSyncSinceID(handle, channelID string, sinceID int64) error {
	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)
	configCol := db.C("config")

	config := BotConfig{}
	err := configCol.Find(bson.M{}).One(&config)
	if err != nil {
		return err
	}

	for i, ts := range config.TweetSyncChannels {
		if ts.Handle == handle && ts.ChannelID == channelID {
			config.TweetSyncChannels[i].SinceID = sinceID
		}
	}

	err = configCol.Update(bson.M{}, bson.M{"$set": bson.M{"tweet_sync_channels": config.TweetSyncChannels}})
	if err != nil {
		return err
	}

	TweetSyncChannels = config.TweetSyncChannels
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

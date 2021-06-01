/*
 * Realtime config from the DB
 */

package config

import (
	"context"
	"errors"
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

var StaffRoles = map[string][]string{
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
	Fanbox  string `json:"fanbox" bson:"fanbox"`
	Former  string `json:"former" bson:"former"`
	Mute    string `json:"mute" bson:"mute"`
}

type TweetSyncConfig struct {
	Handle           string `json:"handle" bson:"handle"`
	ChannelID        string `json:"channel_id" bson:"channel_id"`
	ControlChannelID string `json:"control_channel_id" bson:"control_channel_id"`
	SinceID          int64  `json:"since_id" bson:"since_id"`
}

type TweetUpdate struct {
	ChannelID      string `json:"channel_id" bson:"channel_id"`
	UserMessageID  string `json:"user_message_id" bson:"user_message_id"`
	BotMessageID   string `json:"bot_message_id" bson:"bot_message_id"`
	TweetMessageID string `json:"tweet_message_id" bson:"tweet_message_id"`
	Translation    string `json:"translation" bson:"translation"`
	Translator     string `json:"translator" bson:"translator"`
}

type Extraction struct {
	ChannelID         string   `json:"channel_id" bson:"channel_id"`
	UserMessageID     string   `json:"user_message_id" bson:"user_message_id"`
	BotMessageID      string   `json:"bot_message_id" bson:"bot_message_id"`
	ExtractMessageIDs []string `json:"extract_message_ids" bson:"extract_message_ids"`
}

type CopyPipeline struct {
	OID       bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`

	CreatedBy      string `json:"created_by" bson:"created_by"`
	CreatedByName  string `json:"created_by_name" bson:"created_by_name"`
	Type           string `json:"type" bson:"type"`
	ChannelID      string `json:"channel_id" bson:"channel_id"`
	Prefix         string `json:"prefix" bson:"prefix"`
	YoutubeVideoID string `json:"youtube_video_id" bson:"youtube_video_id"`

	YoutubeVideoTitle string `json:"youtube_video_title" bson:"youtube_video_title"`
	YoutubeLivechatID string `json:"youtube_livechat_id" bson:"youtube_livechat_id"`
}

var GrantRoles = map[string]RoleConfig{
	"755437328515989564": { // DFS
		Alpha:   "760705266953355295",
		Special: "783112023570513970",
		Whale:   "761570574794489886",
		Fanbox:  "",
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
var DateFormat string

var GoogleCredentials bson.M
var GoogleCredentialsAlt1 bson.M
var GoogleOauthCredentials bson.M

var GoogleClientID string
var GoogleSecret string

var YoutubeOauthToken string
var YoutubeRefreshToken string

var EightBallEnabled bool

var Loc *time.Location

var Proposals = make(map[string]string)

var CreatorID = "204752740503650304"

var TweetSyncChannels = []TweetSyncConfig{}

var TweetUpdates = make(map[string]TweetUpdate)

var Extractions = make(map[string]Extraction)

var CopyPipelines = []CopyPipeline{}

var DoubleTL = false

type BotConfig struct {
	ModeratorRoles         map[string][]string   `json:"moderator_roles" bson:"moderator_roles"`
	StaffRoles             map[string][]string   `json:"staff_roles" bson:"staff_roles"`
	GrantRoles             map[string]RoleConfig `json:"grant_roles" bson:"grant_roles"`
	SyncSheets             map[string]string     `json:"sync_sheets" bson:"sync_sheets"`
	RoleGrantEnabled       map[string]bool       `json:"role_grant_enabled" bson:"role_grant_enabled"`
	RoleRemoveEnabled      map[string]bool       `json:"role_remove_enabled" bson:"role_remove_enabled"`
	TimeFormat             string                `json:"time_format" bson:"time_format"`
	DateFormat             string                `json:"date_format" bson:"date_format"`
	GoogleCredentials      bson.M                `json:"-" bson:"google_credentials"`
	GoogleCredentialsAlt1  bson.M                `json:"-" bson:"google_credentials_alt1"`
	GoogleOauthCredentials bson.M                `json:"-" bson:"google_oauth_credentials"`
	GoogleClientID         string                `json:"-" bson:"google_client_id"`
	GoogleSecret           string                `json:"-" bson:"google_secret"`
	YoutubeOauthToken      string                `json:"-" bson:"youtube_oauth_token"`
	YoutubeRefreshToken    string                `json:"-" bson:"youtube_refresh_token"`
	EightBallEnabled       bool                  `json:"eight_ball_enabled" bson:"eight_ball_enabled"`
	TweetSyncChannels      []TweetSyncConfig     `json:"tweet_sync_channels" bson:"tweet_sync_channels"`
	CopyPipelines          []CopyPipeline        `json:"copy_pipelines" bson:"copy_pipelines"`
	DoubleTL               bool                  `json:"double_tl" bson:"double_tl"`
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
	StaffRoles = config.StaffRoles
	GrantRoles = config.GrantRoles
	SyncSheets = config.SyncSheets
	RoleGrantEnabled = config.RoleGrantEnabled
	RoleRemoveEnabled = config.RoleRemoveEnabled
	TimeFormat = config.TimeFormat
	GoogleCredentials = config.GoogleCredentials
	GoogleCredentialsAlt1 = config.GoogleCredentialsAlt1
	GoogleOauthCredentials = config.GoogleOauthCredentials
	GoogleClientID = config.GoogleClientID
	GoogleSecret = config.GoogleSecret
	YoutubeOauthToken = config.YoutubeOauthToken
	YoutubeRefreshToken = config.YoutubeRefreshToken
	EightBallEnabled = config.EightBallEnabled
	TweetSyncChannels = config.TweetSyncChannels
	CopyPipelines = config.CopyPipelines
	DoubleTL = config.DoubleTL

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

// FanboxRole get the designated Fanbox role for the given guild
func FanboxRole(guildID string) string {
	guildRoles, ok := GrantRoles[guildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", guildID)
		return ""
	}

	return guildRoles.Fanbox
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

// MuteRole get the designated muted role for the given guild
func MuteRole(guildID string) string {
	guildRoles, ok := GrantRoles[guildID]
	if !ok {
		log.Printf("Could not find guild roles, %s", guildID)
		return ""
	}

	return guildRoles.Mute
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

func SetFanboxRole(guildID, roleID string) error {
	key := fmt.Sprintf("grant_roles.%s.fanbox", guildID)
	update := bson.M{
		key: roleID,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	if v, ok := GrantRoles[guildID]; ok {
		v.Fanbox = roleID
		GrantRoles[guildID] = v
	} else {
		GrantRoles[guildID] = RoleConfig{
			Fanbox: roleID,
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

func SetMuteRole(guildID, roleID string) error {
	key := fmt.Sprintf("grant_roles.%s.mute", guildID)
	update := bson.M{
		key: roleID,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	if v, ok := GrantRoles[guildID]; ok {
		v.Mute = roleID
		GrantRoles[guildID] = v
	} else {
		GrantRoles[guildID] = RoleConfig{
			Mute: roleID,
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

func SetDoubleTLEnabled(enabled bool) error {
	update := bson.M{
		"double_tl": enabled,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	DoubleTL = enabled
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

func MaybeGetTweetConfig(channelID string) *TweetSyncConfig {
	for _, c := range TweetSyncChannels {
		if c.ControlChannelID == channelID {
			return &c
		}
	}

	return nil
}

func SetCopyPipeline(cp CopyPipeline) error {
	if GetCopyPipeline(cp.ChannelID, cp.YoutubeVideoID) != nil {
		return errors.New("Already copying to that chat")
	}

	if !cp.OID.Valid() {
		cp.OID = bson.NewObjectId()
	}

	cp.CreatedAt = time.Now()

	newCPs := append(CopyPipelines, cp)

	update := bson.M{
		"copy_pipelines": newCPs,
	}

	err := UpdateConfig(update)
	if err != nil {
		return err
	}

	CopyPipelines = newCPs
	return nil
}

func RemoveCopyPipeline(channelID string) ([]CopyPipeline, error) {
	res := []CopyPipeline{}
	removed := []CopyPipeline{}
	for i, cp := range CopyPipelines {
		if cp.ChannelID == channelID {
			removed = append(removed, CopyPipelines[i])
		} else {
			res = append(res, CopyPipelines[i])
		}
	}

	update := bson.M{
		"copy_pipelines": res,
	}

	err := UpdateConfig(update)
	if err != nil {
		return []CopyPipeline{}, err
	}

	CopyPipelines = res
	return removed, nil
}

func GetCopyPipeline(channelID, videoID string) *CopyPipeline {
	for _, cp := range CopyPipelines {
		if cp.ChannelID == channelID && cp.YoutubeVideoID == videoID {
			return &cp
		}
	}

	return nil
}

func GetCopyPipelines(channelID string) []CopyPipeline {
	res := []CopyPipeline{}
	for i, cp := range CopyPipelines {
		if cp.ChannelID == channelID {
			res = append(res, CopyPipelines[i])
		}
	}

	return res
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

func PrintDate(t time.Time) string {
	return t.Format(DateFormat)
}

func Now() time.Time {
	return time.Now().In(Loc)
}

func MessageLink(msg *discordgo.Message) string {
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", msg.GuildID, msg.ChannelID, msg.ID)
}

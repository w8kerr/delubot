/*
 * Models for the app's internal objects
 */

package models

import (
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/globalsign/mgo/bson"
)

// Config Config model
type Config struct {
	SuppressFanboxSweep bool     `bson:"suppress_fanbox_sweep"`
	GoogleCredentials   bson.M   `bson:"google_credentials"`
	AdminTwitterIDs     []string `bson:"admin_twitter_ids"`
	IdolTwitterIDs      []string `bson:"idol_twitter_ids"`
}

// TwitterIDIsAdmin Checks the config object for whether the given Twitter ID should be an Admin
func (con *Config) TwitterIDIsAdmin(twitterID string) bool {
	for _, id := range con.AdminTwitterIDs {
		if twitterID == id {
			return true
		}
	}

	return false
}

// TwitterIDIsIdol Checks the config object for whether the given Twitter ID should be an Idol
func (con *Config) TwitterIDIsIdol(twitterID string) bool {
	for _, id := range con.IdolTwitterIDs {
		if twitterID == id {
			return true
		}
	}

	return false
}

// User User model
type User struct {
	OID bson.ObjectId `json:"_id" bson:"_id,omitempty"`

	PixivUserID   string `json:"pixiv_user_id" bson:"pixiv_user_id"`
	PixivUserName string `json:"pixiv_user_name" bson:"pixiv_user_name"`
	PixivIconURL  string `json:"pixiv_icon_url" bson:"pixiv_icon_url"`

	TwitterUserID  string `json:"twitter_user_id" bson:"twitter_user_id"`
	TwitterHandle  string `json:"twitter_handle" bson:"twitter_handle"`
	TwitterName    string `json:"twitter_name" bson:"twitter_name"`
	TwitterIconURL string `json:"twitter_icon_url" bson:"twitter_icon_url"`

	DiscordUserID        string `json:"discord_user_id" bson:"discord_user_id"`
	DiscordHandle        string `json:"discord_handle" bson:"discord_handle"`
	DiscordDiscriminator string `json:"discord_discriminator" bson:"discord_discriminator"`
	DiscordIconURL       string `json:"discord_icon_url" bson:"discord_icon_url"`
	DiscordNickName      string `json:"discord_nickname" bson:"discord_nickname"`

	GoogleUserID   string `json:"google_user_id" bson:"google_user_id"`
	YoutubeName    string `json:"youtube_name" bson:"youtube_name"`
	YoutubeIconURL string `json:"youtube_icon_url" bson:"youtube_icon_url"`

	IsAdmin bool `json:"is_admin" bson:"is_admin"`
	IsIdol  bool `json:"is_idol" bson:"is_idol"`
}

// Merge Merge all fields (except _id) from the parameter user onto the receiver user
func (u *User) Merge(from User) {
	u.PixivUserID = from.PixivUserID
	u.TwitterUserID = from.TwitterUserID
	u.TwitterHandle = from.TwitterHandle
	u.TwitterName = from.TwitterName
	u.TwitterIconURL = from.TwitterIconURL
	u.DiscordUserID = from.DiscordUserID
	u.DiscordHandle = from.DiscordHandle
	u.DiscordDiscriminator = from.DiscordDiscriminator
	u.DiscordIconURL = from.DiscordIconURL
	u.GoogleUserID = from.GoogleUserID
	u.YoutubeName = from.YoutubeName
	u.YoutubeIconURL = from.YoutubeIconURL
	u.IsAdmin = from.IsAdmin
	u.IsIdol = from.IsIdol
}

// Session Session model
type Session struct {
	OID bson.ObjectId `json:"_id" bson:"_id,omitempty"`

	UserID    bson.ObjectId `json:"user_id" bson:"user_id"`
	AuthToken string        `json:"auth_token" bson:"auth_token"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

const (
	PlatformTwitter = "twitter"
	PlatformDiscord = "discord"
	PlatformGoogle  = "google"
)

// AccessKey Store access key in a collection that won't be sent to the client
type AccessKey struct {
	OID bson.ObjectId `json:"_id" bson:"_id,omitempty"`

	Platform          string `json:"platform" bson:"platform"`
	PlatformUserID    string `json:"platform_user_id" bson:"platform_user_id"`
	AccessToken       string `json:"access_token" bson:"access_token"`
	AccessTokenSecret string `json:"access_token_secret" bson:"access_token_secret"`
}

// CommentReference post ID + comment ID, the minimum info to identify a comment
type CommentReference struct {
	PostID    string `json:"post_id" bson:"post_id"`
	CommentID string `json:"comment_id" bson:"comment_id"`
}

// Verification record of a parsed verification comment
type Verification struct {
	OID           bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	PixivUserID   string        `json:"pixiv_user_id" bson:"pixiv_user_id"`
	PixivUserName string        `json:"pixiv_user_name" bson:"pixiv_user_name"`
	PixivIconURL  string        `json:"pixiv_icon_url" bson:"pixiv_icon_url"`

	TwitterHandle  string        `json:"twitter_handle" bson:"twitter_handle"`
	TwitterComment FanboxComment `json:"twitter_comment" bson:"twitter_comment"`

	DiscordHandle        string        `json:"discord_handle" bson:"discord_handle"`
	DiscordDiscriminator string        `json:"discord_discriminator" bson:"discord_discriminator"`
	DiscordComment       FanboxComment `json:"discord_comment" bson:"discord_comment"`

	UpdateLog []UpdateEvent `json:"update_log" bson:"update_log"`
}

// UpdateEvent record of a verification being updated beyond its original definition
// (because of a change of handle that is matched to an underlying user ID)
type UpdateEvent struct {
	UpdatedAt      time.Time `json:"updated_at" bson:"updated_at"`
	Platform       string    `json:"platform" bson:"platform"`
	PreviousHandle string    `json:"previous_handle" bson:"previous_handle"`
	NewHandle      string    `json:"new_handle" bson:"new_handle"`
	UserID         string    `json:"user_id" bson:"user_id"`
}

// Types of verification events
const (
	EventTypeVerify     = "verify"
	EventTypeInvalidate = "invalidate"
)

// TwitterVerificationEvent Event representing a change in Twitter verification
type TwitterVerificationEvent struct {
	OID           bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	FiredAt       time.Time     `json:"fired_at" bson:"fired_at"`
	Type          string        `json:"type" bson:"type"`
	PixivHandle   string        `json:"pixiv_handle" bson:"pixiv_handle"`
	TwitterHandle string        `json:"twitter_handle" bson:"twitter_handle"`
	Processed     bool          `json:"processed" bson:"processed"`
}

// DiscordVerificationEvent Event representing a change in Discord verification
type DiscordVerificationEvent struct {
	OID                  bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	FiredAt              time.Time     `json:"fired_at" bson:"fired_at"`
	Type                 string        `json:"type" bson:"type"`
	PixivHandle          string        `json:"pixiv_handle" bson:"pixiv_handle"`
	DiscordHandle        string        `json:"discord_hanle" bson:"discord_handle"`
	DiscordDiscriminator string        `json:"discord_discriminator" bson:"discord_discriminator"`
	Processed            bool          `json:"processed" bson:"processed"`
}

// YoutubeStreamRecord Record of an upcoming or finished Youtube stream, parsed from Fanbox posts
type YoutubeStreamRecord struct {
	OID             bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	PostTitle       string        `json:"post_title" bson:"post_title"`
	PostLink        string        `json:"post_link" bson:"post_link"`
	PostPlan        int           `json:"post_plan" bson:"post_plan"`
	YoutubeID       string        `json:"youtube_id" bson:"youtube_id"`
	Completed       bool          `json:"completed" bson:"completed"`
	ScheduledTime   time.Time     `json:"scheduled_time" bson:"scheduled_time"`
	StreamTitle     string        `json:"stream_title" bson:"stream_title"`
	StreamThumbnail string        `json:"stream_thumbnail" bson:"stream_thumbnail"`
}

// SyncedTweet Record of a tweet that was echoed from Twitter into Discord
type SyncedTweet struct {
	OID              bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	ChannelID        string        `json:"channel_id" bson:"channel_id"`
	MessageID        string        `json:"message_id" bson:"message_id"`
	ControlChannelID string        `json:"control_channel_id" bson:"control_channel_id"`
	ControlMessageID string        `json:"control_message_id" bson:"control_message_id"`
	Tweet            twitter.Tweet `json:"tweet" bson:"tweet"`
	Translation      string        `json:"translation" bson:"translation"`
	CreatedAt        time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" bson:"updated_at"`
	Translators      []string      `json:"translators" bson:"translators"`
	HumanTranslated  bool          `json:"human_translated" bson:"human_translated"`
}

// WatchedVideo Record a video that should be scanned for new comments
type WatchedVideo struct {
	OID       bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	VideoID   string        `json:"video_id" bson:"video_id"`
	LastScan  time.Time     `json:"last_scan" bson:"last_scan"`
	ChannelID string        `json:"channel_id" bson:"channel_id"`
}

type YoutubeComment struct {
	AuthorDisplayName     string
	AuthorProfileImageURL string
	Text                  string
	ReplyDisplayName      string
	ReplyText             string
	UpdatedAt             time.Time
}

type Sticky struct {
	OID             bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	AuthorName      string        `json:"author_name" bson:"author_name"`
	AuthorAvatarURL string        `json:"author_avatar_url" bson:"author_avatar_url"`
	ChannelID       string        `json:"channel_id" bson:"channel_id"`
	MessageID       string        `json:"message_id" bson:"message_id"`
	Text            string        `json:"text" bson:"text"`
	Time            time.Time     `json:"time" bson:"time"`
}

// BTableOptions holds metadata about how to sort and paginate a query for a
// Bootstrap-Vue table provider
type BTableOptions struct {
	Filter      string
	PageSize    int
	CurrentPage int
	Skip        int
	Exclude     []bson.ObjectId
	ExcludeStr  []string
	SortBy      string
	SortDir     int
	Active      bool
}

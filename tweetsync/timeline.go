package tweetsync

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/globalsign/mgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/models"
	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/tl"
	"github.com/w8kerr/delubot/utils"
)

var TwitterTimeFormat = "Mon Jan 2 15:04:05 +0000 2006"

// InitTimelines initialize all streams for Tweet streaming
func InitTimelines(ds *discordgo.Session) {
	fmt.Println("InitTimelines")
	if len(config.TweetSyncChannels) == 0 {
		return
	}

	apiKey := os.Getenv("TWITTER_API_KEY")
	apiSecret := os.Getenv("TWITTER_API_SECRET")

	userToken := os.Getenv("DELU_TWEETSYNC_TOKEN")
	userSecret := os.Getenv("DELU_TWEETSYNC_SECRET")

	con := oauth1.NewConfig(apiKey, apiSecret)
	token := oauth1.NewToken(userToken, userSecret)
	httpClient := con.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	for i := range config.TweetSyncChannels {
		ScanTimeline(ds, client, &config.TweetSyncChannels[i])
	}
}

// ScanTimeline polls a stream of Tweets and posts them in the specified channel
func ScanTimeline(ds *discordgo.Session, tc *twitter.Client, ts *config.TweetSyncConfig) {
	fmt.Println("Init Tweetsync - Handle", ts.Handle)
	fmt.Println("Init Tweetsync - Channel ID", ts.ChannelID)

	// Don't return any tweets, just set the most recent one
	if ts.SinceID == 0 {
		trimUser := true
		tweets, _, err := tc.Timelines.UserTimeline(&twitter.UserTimelineParams{
			ScreenName: ts.Handle,
			Count:      1,
			TrimUser:   &trimUser,
		})
		if err != nil {
			log.Printf("Failed to initialize Tweet stream, %s", err)
			return
		}
		if len(tweets) == 0 {
			log.Printf("Failed to initialize Tweet stream, no tweets returns")
			return
		}

		fmt.Println("Set most recent tweet", ts.Handle, ts.ChannelID, tweets[0].IDStr)
		config.SetTweetSyncSinceID(ts.Handle, ts.ChannelID, tweets[0].ID)
		return
	}

	go Scan(ds, tc, ts)
}

func Scan(ds *discordgo.Session, tc *twitter.Client, ts *config.TweetSyncConfig) {
	sleepDuration := 3 * time.Second
	sinceID := ts.SinceID

	for {
		time.Sleep(sleepDuration)

		tweets, _, err := tc.Timelines.UserTimeline(&twitter.UserTimelineParams{
			ScreenName: ts.Handle,
			SinceID:    sinceID,
		})
		if err != nil {
			log.Printf("Failed to initialize Tweet stream, %s", err)
			return
		}

		sort.Slice(tweets, func(i, j int) bool {
			return tweets[i].ID < tweets[j].ID
		})

		for _, tweet := range tweets {
			translation, err := tl.Translate(tweet.Text)
			if err != nil {
				translation = fmt.Sprintf("[Translation error: %s]", err)
			}

			st := models.SyncedTweet{
				Tweet:           tweet,
				Translation:     translation,
				CreatedAt:       time.Now(),
				Translators:     []string{"Google Translate"},
				HumanTranslated: false,
			}

			embed := SyncedTweetToEmbed(st)
			msg, err := ds.ChannelMessageSendEmbed(ts.ChannelID, embed)
			if err != nil {
				log.Printf("Failed to send Tweet %s, %s", tweet.IDStr, err)
				return
			}
			st.MessageID = msg.ID

			// Save to the DB
			session := mongo.MDB.Clone()
			defer session.Close()
			session.SetMode(mgo.Strong, false)
			db := session.DB(mongo.DB_NAME)
			stCol := db.C("synced_tweets")
			stCol.Insert(st)

			sinceID = tweets[0].ID
			config.SetTweetSyncSinceID(ts.Handle, ts.ChannelID, tweets[0].ID)
		}
	}
}

func SyncedTweetToEmbed(st models.SyncedTweet) *discordgo.MessageEmbed {
	return TweetToEmbed(&st.Tweet, st.Translation, st.Translators)
}

func TweetToEmbed(tweet *twitter.Tweet, translation string, translators []string) *discordgo.MessageEmbed {
	createdAt, _ := time.Parse(TwitterTimeFormat, tweet.CreatedAt)

	fmt.Println("TWEET")
	utils.PrintJSON(tweet)

	// Special Twitter action, make the profile picture bigger
	parts := strings.Split(tweet.User.ProfileImageURLHttps, "_normal.")
	if len(parts) == 2 {
		tweet.User.ProfileImageURLHttps = parts[0] + "_400x400." + parts[1]
	}

	sourceParts := strings.Split(tweet.Source, "\u003e")
	if len(sourceParts) > 2 {
		sourceParts2 := strings.Split(sourceParts[1], "\u003c")
		tweet.Source = sourceParts2[0]
	}

	embed := &discordgo.MessageEmbed{
		Color: 3066993,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: tweet.User.ProfileImageURLHttps,
		},
		Author: &discordgo.MessageEmbedAuthor{
			Name: fmt.Sprintf("%s\n@%s", tweet.User.Name, tweet.User.ScreenName),
			URL:  fmt.Sprintf("https://twitter.com/%s", tweet.User.ScreenName),
		},
		Description: fmt.Sprintf("[Status: %s](https://twitter.com/%s/%s)\n────────────────────", tweet.IDStr, tweet.User.ScreenName, tweet.IDStr),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  fmt.Sprintf("❝ %s ❞", translation),
				Value: fmt.Sprintf("\n\u200B\n*TL: %s*\n────────────────────", strings.Join(translators, ", ")),
			},
			{
				Name:  fmt.Sprintf("❝ %s ❞", tweet.Text),
				Value: "\n\u200B\n*Original*\n────────────────────",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: tweet.Source,
		},
		Timestamp: createdAt.Format(time.RFC3339),
	}

	return embed
}

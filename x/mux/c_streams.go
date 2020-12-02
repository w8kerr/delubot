package mux

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/w8kerr/delubot/models"
	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/youtubesvc"
)

func (m *Mux) Streams(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	prerespond := GetResponder(ds, dm)
	msg := prerespond("🔺Looking up stream information...")
	respond := GetEditor(ds, msg)

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	c := context.Background()
	c = context.WithValue(c, "mgo", session)
	db := session.DB(mongo.DB_NAME)

	ytCol := db.C("youtube_stream_records")
	recs := []models.YoutubeStreamRecord{}
	err := ytCol.Find(bson.M{"completed": false}).All(&recs)
	if err != nil {
		respond("🔺Could not get stream information, " + err.Error())
		return
	}

	if len(recs) == 0 {
		respond("🔺No upcoming streams found :(")
		return
	}

	ytSvc, err := youtubesvc.NewYoutubeService(c)
	if err != nil {
		respond("🔺Could not connect to Youtube")
		return
	}
	for i, rec := range recs {
		scheduledTime, _, snippet, err := ytSvc.GetStreamInfo(rec.YoutubeID)
		if err != nil {
			respond("🔺Error getting video info: " + err.Error())
		}
		recs[i].ScheduledTime = scheduledTime
		recs[i].StreamTitle = snippet.Title
		recs[i].StreamThumbnail = snippet.Thumbnails.High.Url
	}

	sort.Slice(recs, func(a int, b int) bool {
		return recs[a].ScheduledTime.Before(recs[b].ScheduledTime)
	})

	embeds := []*discordgo.MessageEmbed{}
	for _, rec := range recs {
		embed := &discordgo.MessageEmbed{
			Title:       rec.StreamTitle,
			Description: fmt.Sprintf("%s\nSee %s for link", TimeBefore(rec.ScheduledTime), rec.PostLink),
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("\n`Restricted to ¥%d plan members`", rec.PostPlan),
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: rec.StreamThumbnail,
			},
			Color: 3066993,
		}
		embeds = append(embeds, embed)
		// resp := fmt.Sprintf("> **%s**", rec.StreamTitle)
		// resp += fmt.Sprintf("\n%s", TimeBefore(rec.ScheduledTime))
		// resp += fmt.Sprintf("\nSee %s for link", rec.PostLink)
		// if rec.PostPlan > 500 {
		// 	resp += fmt.Sprintf("\n`Restricted to ¥%d plan members`", rec.PostPlan)
		// }

		// resps = append(resps, resp)
	}

	final := "🔺Upcoming streams:"
	respond(final)
	for _, embed := range embeds {
		ds.ChannelMessageSendEmbed(dm.ChannelID, embed)
	}
}

func TimeBefore(t time.Time) string {
	currentTime := time.Now()
	difference := t.Sub(currentTime)

	total := int(difference.Seconds())
	days := int(total / (60 * 60 * 24))
	hours := int(total / (60 * 60) % 24)
	minutes := int(total/60) % 60

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d days", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hours", hours))
	}
	parts = append(parts, fmt.Sprintf("%d minutes", minutes))

	return strings.Join(parts, ", ") + " from now"
}

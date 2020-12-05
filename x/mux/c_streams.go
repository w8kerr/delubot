package mux

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/w8kerr/delubot/config"
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

	if len(recs) > 0 {
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
					Text: fmt.Sprintf("\nRestricted to ¥%d plan members", rec.PostPlan),
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

	schedCol := db.C("scheduled_streams")
	schedStreams := []ManualStream{}
	err = schedCol.Find(bson.M{"time": bson.M{"$gt": time.Now()}}).Sort("time").All(&schedStreams)
	if err != nil && err != mgo.ErrNotFound {
		respond("Failed to get manually scheduled streams")
	}

	if len(schedStreams) > 0 {
		if len(recs) > 0 {
			prerespond("🔺Upcoming scheduled streams:")
		} else {
			respond("🔺Upcoming scheduled streams:")
		}
		for _, schedStream := range schedStreams {
			collision := false
			for _, rec := range recs {
				diff := schedStream.Time.Sub(rec.ScheduledTime).Hours()
				if diff > -1.1 && diff < 1.1 {
					collision = true
				}
			}

			if !collision {
				embed := &discordgo.MessageEmbed{
					Title:       schedStream.Title,
					Description: TimeBefore(schedStream.Time),
					Color:       10181046,
				}
				ds.ChannelMessageSendEmbed(dm.ChannelID, embed)
			}
		}
	}

	if len(recs) == 0 && len(schedStreams) == 0 {
		respond("🔺No upcoming streams found :(")
		return
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

type ManualStream struct {
	Time  time.Time `json:"time" bson:"time"`
	Title string    `json:"title" bson:"title"`
}

var addStreamRE = regexp.MustCompile(`(\d\d\d\d\/\d\d\/\d\d \d\d:\d\d) (.+)`)

func (m *Mux) AddStream(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	cmd := strings.TrimSpace(strings.TrimPrefix(ctx.Content, "addstream"))

	match := addStreamRE.FindAllSubmatch([]byte(cmd), -1)
	if match == nil {
		respond("🔺Usage: `-addstream yyyy/mm/dd hh:mm <title>`")
		return
	}

	timeStr := string(match[0][1])
	titleStr := string(match[0][2])

	Loc, _ := time.LoadLocation("Asia/Tokyo")
	t, err := time.ParseInLocation("06/01/02 15:04", timeStr, Loc)
	if err != nil {
		respond("🔺I don't understand that stream time :(\nUsage: `-addstream yyyy/mm/dd hh:mm <title>` (" + err.Error() + ")")
		return
	}

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	c := context.Background()
	c = context.WithValue(c, "mgo", session)
	db := session.DB(mongo.DB_NAME)

	stream := ManualStream{
		Time:  t,
		Title: titleStr,
	}
	schedCol := db.C("scheduled_streams")
	schedCol.Upsert(bson.M{"time": stream.Time}, stream)
	respond("🔺Stream added at " + config.PrintTime(t))
}

var removeStreamRE = regexp.MustCompile(`(\d\d\d\d\/\d\d\/\d\d \d\d:\d\d)`)

func (m *Mux) RemoveStream(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	cmd := strings.TrimSpace(strings.TrimPrefix(ctx.Content, "removestream"))

	match := removeStreamRE.FindAllSubmatch([]byte(cmd), -1)
	if match == nil {
		respond("🔺Usage: `-removestream yyyy/mm/dd hh:mm`")
		return
	}

	timeStr := string(match[0][1])

	Loc, _ := time.LoadLocation("Asia/Tokyo")
	t, err := time.ParseInLocation("06/01/02 15:04", timeStr, Loc)
	if err != nil {
		respond("🔺I don't understand that stream time :(\nUsage: `-removestream yyyy/mm/dd hh:mm` (" + err.Error() + ")")
		return
	}

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	c := context.Background()
	c = context.WithValue(c, "mgo", session)
	db := session.DB(mongo.DB_NAME)

	schedCol := db.C("scheduled_streams")
	stream := ManualStream{}
	err = schedCol.Find(bson.M{"time": t}).One(&stream)
	if err != nil {
		respond("🔺I couldn't find a stream at " + config.PrintTime(t) + " :(")
	} else {
		schedCol.Remove(bson.M{"time": t})
		respond("🔺Stream removed")
	}
}

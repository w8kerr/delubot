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

type EmbedToUpdate struct {
	ChannelID string
	MessageID string
	Time      time.Time
}

var EmbedsToUpdate = []EmbedToUpdate{}

func (m *Mux) Stream(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)
	respond("🔺No fuck you it's supposed to be 'streams' >:l")
}

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

	ytSvc, err := youtubesvc.NewYoutubeService(c)
	if err != nil {
		respond("🔺Could not connect to Youtube, " + err.Error())
		return
	}

	if len(recs) > 0 {
		for i, rec := range recs {
			scheduledTime, _, snippet, err := ytSvc.GetStreamInfo(rec.YoutubeID)
			if err != nil {
				respond("🔺Error getting video info: " + err.Error())
			}
			recs[i].ScheduledTime = scheduledTime
			recs[i].StreamTitle = snippet.Title
			recs[i].StreamThumbnail = snippet.Thumbnails.High.Url
		}
	}

	liveRecs, err := ytSvc.ListUpcomingStreams("UC7YXqPO3eUnxbJ6rN0z2z1Q")
	if err != nil {
		respond("🔺Could not check upcoming Youtube streams, " + err.Error())
		return
	}

	recs = append(liveRecs, recs...)

	schedCol := db.C("scheduled_streams")
	schedStreams := []ManualStream{}
	err = schedCol.Find(bson.M{"time": bson.M{"$gt": time.Now()}}).Sort("time").All(&schedStreams)
	if err != nil && err != mgo.ErrNotFound {
		respond("🔺Failed to get manually scheduled streams: " + err.Error())
	}

	if len(schedStreams) > 0 {
		Loc, _ := time.LoadLocation("Asia/Tokyo")
		for _, schedStream := range schedStreams {
			schedStream.Time = schedStream.Time.In(Loc)
		}
	}

	if len(recs) == 0 && len(schedStreams) == 0 {
		respond("🔺No upcoming streams found :(")
		return
	}

	embed := StreamsEmbed(schedStreams, recs)

	ds.ChannelMessageDelete(dm.ChannelID, msg.ID)
	msg, err = ds.ChannelMessageSendEmbed(dm.ChannelID, embed)
	if err == nil {
		EmbedsToUpdate = append(EmbedsToUpdate, EmbedToUpdate{
			ChannelID: dm.ChannelID,
			MessageID: msg.ID,
			Time:      time.Now(),
		})
	}
}

func StreamsEmbed(mans []ManualStream, recs []models.YoutubeStreamRecord) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color: 3066993,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Updated",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	sort.Slice(recs, func(a int, b int) bool {
		return recs[a].ScheduledTime.Before(recs[b].ScheduledTime)
	})

	Loc, _ := time.LoadLocation("Asia/Tokyo")

	fields := []*discordgo.MessageEmbedField{}

	for _, rec := range recs {
		rec.ScheduledTime = rec.ScheduledTime.In(Loc)
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "🔺" + rec.StreamTitle,
			Value: fmt.Sprintf("[See Fanbox for link](%s)\nRestricted to ¥%d plan members\n%s\n%s", rec.PostLink, rec.PostPlan, TimeBefore(rec.ScheduledTime), config.PrintTime(rec.ScheduledTime)),
		})
	}

	for _, man := range mans {
		if man.ReplacedBy(recs) {
			continue
		}

		if man.GuerrillaTime == "" {
			man.Time = man.Time.In(Loc)
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  "🔺" + man.Title,
				Value: fmt.Sprintf("%s\n%s", TimeBefore(man.Time), config.PrintTime(man.Time)),
			})
		} else {
			man.Time = man.Time.In(Loc)
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  "🔺" + man.Title,
				Value: fmt.Sprintf("%s\n❓%s (%s)", EightHourRange(man.Time), config.PrintDate(man.Time), man.GuerrillaTime),
			})
		}

	}

	if len(recs) > 0 {
		embed.Title = "Upcoming streams:"
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: recs[0].StreamThumbnail,
		}
	} else {
		embed.Title = "Upcoming scheduled streams:"
	}

	embed.Fields = fields

	return embed
}

func TimeBefore(t time.Time) string {
	currentTime := time.Now()
	difference := t.Sub(currentTime)

	total := int(difference.Seconds())
	days := int(total / (60 * 60 * 24))
	hours := int(total / (60 * 60) % 24)
	minutes := int(total/60) % 60
	deltas := float32(total) / (60 * 60 * 4)

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d days", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hours", hours))
	}
	parts = append(parts, fmt.Sprintf("%d minutes", minutes))

	return fmt.Sprintf("%s (%.1f Δs) from now", strings.Join(parts, ", "), deltas)
}

func EightHourRange(t time.Time) string {
	currentTime := time.Now()
	lRange := t.Add(-4 * time.Hour)
	uRange := t.Add(4 * time.Hour)

	if currentTime.After(lRange) {
		return "Any time now!"
	}

	lDiff := lRange.Sub(currentTime)

	lTotal := int(lDiff.Seconds())
	lDays := int(lTotal / (60 * 60 * 24))
	lHours := int(lTotal / (60 * 60) % 24)

	uDiff := uRange.Sub(currentTime)

	uTotal := int(uDiff.Seconds())
	uDays := int(uTotal / (60 * 60 * 24))
	uHours := int(uTotal / (60 * 60) % 24)

	parts := []string{}

	if lDays > 0 {
		if lDays > 1 {
			parts = append(parts, fmt.Sprintf("〜 %d days and", lDays))
		} else {
			parts = append(parts, fmt.Sprintf("〜 %d day and", lDays))
		}
	}
	parts = append(parts, fmt.Sprintf("%d hours", lHours))
	parts = append(parts, "to")
	if uDays > 0 {
		if lDays > 1 {
			parts = append(parts, fmt.Sprintf("%d days and", uDays))
		} else {
			parts = append(parts, fmt.Sprintf("%d day and", uDays))
		}
	}
	parts = append(parts, fmt.Sprintf("%d hours", uHours))

	return fmt.Sprintf("%s from now", strings.Join(parts, " "))
}

type ManualStream struct {
	Time          time.Time `json:"time" bson:"time"`
	Title         string    `json:"title" bson:"title"`
	GuerrillaTime string    `json:"guerrilla_time" bson:"guerrilla_time"`
}

func (ms *ManualStream) ReplacedBy(recs []models.YoutubeStreamRecord) bool {
	for _, rec := range recs {
		diff := ms.Time.Sub(rec.ScheduledTime).Hours()
		if diff > -1.1 && diff < 1.1 {
			return true
		}
	}

	return false
}

var addGuerrillaRE = regexp.MustCompile(`(\d\d\d\d\/\d\d\/\d\d \d\d:\d\d) ([\S]+) (.+)`)

func (m *Mux) AddGuerrilla(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	respond := GetResponder(ds, dm)

	cmd := strings.TrimSpace(strings.TrimPrefix(ctx.Content, "addstream"))

	match := addGuerrillaRE.FindAllSubmatch([]byte(cmd), -1)
	if match == nil {
		respond("🔺Usage: `-addguerrilla yyyy/mm/dd hh:mm <est. time> <title>`")
		return
	}

	timeStr := string(match[0][1])
	guerStr := string(match[0][2])
	titleStr := string(match[0][3])

	Loc, _ := time.LoadLocation("Asia/Tokyo")
	t, err := time.ParseInLocation("2006/01/02 15:04", timeStr, Loc)
	if err != nil {
		respond("🔺I don't understand that stream time :(\nUsage: `-addguerrilla yyyy/mm/dd hh:mm <est. time> <title>` (" + err.Error() + ")")
		return
	}

	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	c := context.Background()
	c = context.WithValue(c, "mgo", session)
	db := session.DB(mongo.DB_NAME)

	stream := ManualStream{
		Time:          t,
		Title:         titleStr,
		GuerrillaTime: guerStr,
	}
	schedCol := db.C("scheduled_streams")
	schedCol.Upsert(bson.M{"time": stream.Time}, stream)
	respond("🔺Guerilla stream added at " + config.PrintDate(t) + " (" + guerStr + ")")
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
	t, err := time.ParseInLocation("2006/01/02 15:04", timeStr, Loc)
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
	t, err := time.ParseInLocation("2006/01/02 15:04", timeStr, Loc)
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

func (m *Mux) InitScanForUpdates(ds *discordgo.Session) {
	sleepDuration := 60 * time.Second
	for {
		time.Sleep(sleepDuration)
		m.ScanForUpdates(ds)
	}
}

func (m *Mux) ScanForUpdates(ds *discordgo.Session) {
	// fmt.Println("SCAN FOR UPDATES")
	pruned := []EmbedToUpdate{}
	for _, etu := range EmbedsToUpdate {
		if etu.Time.Add(24 * time.Hour).After(time.Now()) {
			pruned = append(pruned, etu)
		}
	}
	EmbedsToUpdate = pruned
	if len(EmbedsToUpdate) == 0 {
		return
	}

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
		fmt.Println("🔺Could not get stream information, " + err.Error())
		return
	}

	if len(recs) > 0 {
		ytSvc, err := youtubesvc.NewYoutubeService(c)
		if err != nil {
			fmt.Println("🔺Could not connect to Youtube")
			return
		}
		for i, rec := range recs {
			scheduledTime, _, snippet, err := ytSvc.GetStreamInfo(rec.YoutubeID)
			if err != nil {
				fmt.Println("🔺Error getting video info: " + err.Error())
			}
			recs[i].ScheduledTime = scheduledTime
			recs[i].StreamTitle = snippet.Title
			recs[i].StreamThumbnail = snippet.Thumbnails.High.Url
		}
	}

	schedCol := db.C("scheduled_streams")
	schedStreams := []ManualStream{}
	err = schedCol.Find(bson.M{"time": bson.M{"$gt": time.Now()}}).Sort("time").All(&schedStreams)
	if err != nil && err != mgo.ErrNotFound {
		fmt.Println("Failed to get manually scheduled streams: " + err.Error())
	}

	if len(schedStreams) > 0 {
		Loc, _ := time.LoadLocation("Asia/Tokyo")
		for _, schedStream := range schedStreams {
			schedStream.Time = schedStream.Time.In(Loc)
		}
	}

	if len(recs) == 0 && len(schedStreams) == 0 {
		for _, etu := range EmbedsToUpdate {
			ds.ChannelMessageEdit(etu.ChannelID, etu.MessageID, "🔺No upcoming streams found :(")
		}
		EmbedsToUpdate = []EmbedToUpdate{}
		return
	}

	embed := StreamsEmbed(schedStreams, recs)
	succeeded := []EmbedToUpdate{}
	for _, etu := range EmbedsToUpdate {
		_, err = ds.ChannelMessageEditEmbed(etu.ChannelID, etu.MessageID, embed)
		if err != nil {
			fmt.Println("Failed to update previous streams embed: " + err.Error())
		} else {
			succeeded = append(succeeded, etu)
		}
	}
	EmbedsToUpdate = succeeded
}

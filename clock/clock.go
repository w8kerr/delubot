package clock

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func RunClockChannel(ds *discordgo.Session) {
	for {
		now := time.Now()
		sleepDur := time.Duration(60 - now.Second())

		UpdateClockChannel(ds, now)
		time.Sleep(sleepDur * time.Second)
	}
}

func UpdateClockChannel(ds *discordgo.Session, now time.Time) {
	timeStr := FormatTime(now)

	channelID := "831364073794437170"
	fmt.Println("Update clock channel to", timeStr)
	_, err := ds.ChannelEdit(channelID, timeStr)
	if err != nil {
		log.Printf("Failed to edit channel: %s", err)
	}
}

func RunClockName(ds *discordgo.Session) {
	for {
		now := time.Now()
		sleepDur := time.Duration(60 - now.Second())

		UpdateClockName(ds, now)

		time.Sleep(sleepDur * time.Second)
	}
}

func UpdateClockName(ds *discordgo.Session, now time.Time) {
	timeStr := FormatTime(now)

	guildID := "755437328515989564"
	userID := "@me"
	fmt.Println("Update clock name to", timeStr)
	err := ds.GuildMemberNickname(guildID, userID, timeStr)
	if err != nil {
		log.Printf("Failed to edit nickname: %s", err)
	}
}

func FormatTime(now time.Time) string {
	//init the loc
	loc, _ := time.LoadLocation("Asia/Tokyo")

	//set timezone,
	now = now.In(loc)
	return fmt.Sprintf("%s %s JST", GetClockEmoji(now), now.Format("15:04"))
}

var clockEmojiTop = map[int]string{
	0:  "🕛",
	1:  "🕐",
	2:  "🕑",
	3:  "🕒",
	4:  "🕓",
	5:  "🕔",
	6:  "🕕",
	7:  "🕖",
	8:  "🕗",
	9:  "🕘",
	10: "🕙",
	11: "🕚",
}

var clockEmojiBottom = map[int]string{
	0:  "🕧",
	1:  "🕜",
	2:  "🕝",
	3:  "🕞",
	4:  "🕟",
	5:  "🕠",
	6:  "🕡",
	7:  "🕢",
	8:  "🕣",
	9:  "🕤",
	10: "🕥",
	11: "🕦",
}

func GetClockEmoji(now time.Time) string {
	adj := now.Add(15 * time.Minute)
	if adj.Minute() < 30 {
		emoji := clockEmojiTop[now.Hour()%12]
		return emoji
	} else {
		emoji := clockEmojiBottom[now.Hour()%12]
		return emoji
	}
}

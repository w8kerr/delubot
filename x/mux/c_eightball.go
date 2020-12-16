package mux

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
)

func (m *Mux) EightBall(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	prerespond := GetResponder(ds, dm)

	// emoji, err := ds.State.Emoji(dm.GuildID, "788243303816364062")
	// if err != nil {
	// 	prerespond(fmt.Sprintf("🔺No more eight ball I dropped it on the floor (" + err.Error() + ")"))
	// 	return
	// }

	ctx.Content = strings.TrimPrefix(ctx.Content, "avatar")
	ctx.Content = strings.TrimSpace(ctx.Content)

	if ctx.Content == "enable" && dm.Author.ID == config.CreatorID {
		prerespond("🔺Ah! I dropped the eight ball! " + config.Emoji("delucry"))
		return
	}
	if ctx.Content == "disable" && dm.Author.ID == config.CreatorID {
		prerespond("🔺I found a new eight ball! A listener gave it to me! " + config.Emoji("deluyay"))
		return
	}

	if !config.EightBallEnabled {
		prerespond("🔺No more eight ball I dropped it on the floor " + config.Emoji("notamusedtea"))
		return
	}

	if ctx.Content == "" {
		prerespond("🔺Usage: -db 8ball <yes or no question>")
		return
	}

	answers := []string{
		"It is certain",
		"It is decidedly so",
		"Without a doubt",
		"Yes definitely",
		"You may rely on it",
		"As I see it yes",
		"Most likely",
		"Outlook good",
		"Yes",
		"Signs point to yes",
		"Reply hazy try again",
		"Ask again later",
		"Better not tell you now",
		"Cannot predict now",
		"Concentrate and ask again",
		"Don't count on it",
		"My reply is no",
		"My sources say no",
		"Outlook not so good",
		"Very doubtful",
		"Δ",
		"There are three sides to everything",
	}
	rand.Seed(time.Now().UnixNano())

	msg := prerespond("🔺Picking up the 8 ball...")
	time.Sleep(500 * time.Millisecond)
	respond := GetEditor(ds, msg)

	respond("🔺Shaking the 8 ball...")
	time.Sleep(500 * time.Millisecond)
	respond("🔺Consulting the Triangle Illuminati...")
	time.Sleep(1000 * time.Millisecond)
	respond("🔺Revealing the secrets of the DelUniverse...")
	time.Sleep(1500 * time.Millisecond)
	if rand.Intn(6) == 0 {
		respond("🔺Trying again because I didn't like the answer...")
		time.Sleep(2000 * time.Millisecond)
		respond("🔺Picking up the 8 ball...")
		time.Sleep(500 * time.Millisecond)
		respond("🔺Shaking the 8 ball...")
		time.Sleep(500 * time.Millisecond)
		respond("🔺Consulting the Triangle Illuminati...")
		time.Sleep(1000 * time.Millisecond)
		respond("🔺Revealing the secrets of the DelUniverse...")
		time.Sleep(1500 * time.Millisecond)
	}

	respond(fmt.Sprintf("🔺DeluBot says: ```%s```", answers[rand.Intn(len(answers))]))
}

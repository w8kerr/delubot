package youtubesvc

import (
	"context"
	"fmt"
	"testing"

	"github.com/globalsign/mgo"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/utils"
)

func Test_VideoID(t *testing.T) {
	ctx := context.Background()
	svc, err := NewYoutubeService(ctx)
	if err != nil {
		fmt.Println("Error", err)
	}

	videoID1, err := svc.ParseVideoID("https://www.youtube.com/watch?v=1Mm2VgxI-nA")
	if err != nil {
		fmt.Println("Error", err)
	}
	fmt.Println("1:", videoID1)

	videoID2, err := svc.ParseVideoID("1Mm2VgxI-nA")
	if err != nil {
		fmt.Println("Error", err)
	}
	fmt.Println("2:", videoID2)
}

func Test_LiveChatID(t *testing.T) {
	mongo.Init(false)
	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	c := context.Background()
	c = context.WithValue(c, "mgo", session)

	config.LoadConfig()

	svc, err := NewYoutubeService(c)
	if err != nil {
		fmt.Println("Error", err)
	}

	id, title, err := svc.GetLivechatID("a7I2yMMGe3w")
	if err != nil {
		fmt.Println("Error", err)
	}
	fmt.Println(title, id)
}

func Test_SendChatMessage(t *testing.T) {
	mongo.Init(false)
	session := mongo.MDB.Clone()
	defer session.Close()
	session.SetMode(mgo.Strong, false)
	c := context.Background()
	c = context.WithValue(c, "mgo", session)

	config.LoadConfig()

	svc, err := NewUserYoutubeService(config.YoutubeCredentials[0].OauthToken, &config.YoutubeCredentials[0].RefreshToken)
	if err != nil {
		fmt.Println("Error", err)
	}

	liveChatID := "Cg0KC2E3STJ5TU1HZTN3KicKGFVDc0FKQnl1QzBTa2U2NG1DTjRPRHNhQRILYTdJMnlNTUdlM3c"

	sent, err := svc.SendChatMessage(liveChatID, "I am the one who chats")
	utils.PrintJSON(sent)
}

package youtubesvc

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
	"google.golang.org/api/youtube/v3"
)

var SS *YoutubeService
var DS *discordgo.Session

func InitSweeper(ds *discordgo.Session) {
	DS = ds
	ctx := context.Background()
	var err error
	SS, err = NewYoutubeService(ctx)
	if err != nil {
		log.Printf("Failed to initialize Youtube service")
		return
	}
}

func Sweeper() {
	clog := utils.GetChannelLogger(DS, "793361959046217778")
	sleepDuration := 500 * time.Second
	time.Sleep(3 * time.Second)
	for {
		session := mongo.MDB.Clone()
		defer session.Close()
		session.SetMode(mgo.Strong, false)
		db := session.DB(mongo.DB_NAME)
		wvCol := db.C("watched_videos")

		wvs := []models.WatchedVideo{}
		err := wvCol.Find(bson.M{}).All(&wvs)
		if err != nil {
			log.Printf("Failed to get watched videos")
		}

		for _, wv := range wvs {
			start := time.Now()
			Scan(wv)
			err := wvCol.UpdateId(wv.OID, bson.M{"$set": bson.M{"last_scan": start}})
			if err != nil {
				log.Println("Failed to update last scan time:", err)
				clog.Println("Failed to update last scan time:", err)
				return
			}
		}

		time.Sleep(sleepDuration)
	}
}

func Scan(wv models.WatchedVideo) {
	video, err := SS.service.Videos.List([]string{"snippet"}).Id(wv.VideoID).Do()
	if err != nil {
		log.Println("Failed to get video:", err)
		return
	}
	videoTitle := video.Items[0].Snippet.Title

	now := time.Now()
	comments, err := GetAllComments(wv.VideoID)
	if err != nil {
		log.Println("Failed to get comment threads:", err)
		return
	}
	fmt.Println("Scraped", len(comments), "in", time.Now().Sub(now))

	for _, c := range comments {
		if c.UpdatedAt.After(wv.LastScan) {
			embed := YoutubeCommentToEmbed(c, wv.VideoID, videoTitle)
			_, err := DS.ChannelMessageSendEmbed(wv.ChannelID, embed)
			if err != nil {
				log.Println("Failed to send message: %s", err)
			}
		}
	}
}

func YoutubeCommentToEmbed(c models.YoutubeComment, videoID, videoTitle string) *discordgo.MessageEmbed {
	desc := c.Text

	if c.ReplyText != "" {
		desc += fmt.Sprintf("\n──────────────\nIn reply to:\n%s\n\n%s", c.ReplyDisplayName, c.ReplyText)
	}

	embed := &discordgo.MessageEmbed{
		Color: 3066993,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: c.AuthorProfileImageURL,
		},
		Author: &discordgo.MessageEmbedAuthor{
			Name: c.AuthorDisplayName,
			URL:  "https://www.youtube.com/watch?v=" + videoID,
		},
		Description: desc,
		Footer: &discordgo.MessageEmbedFooter{
			Text: videoTitle,
		},
		Timestamp: c.UpdatedAt.Format(time.RFC3339),
	}

	return embed
}

func ProcessCommentThread(ct *youtube.CommentThread, comments *[]models.YoutubeComment) error {
	t, _ := time.Parse(time.RFC3339, ct.Snippet.TopLevelComment.Snippet.UpdatedAt)
	*comments = append(*comments, models.YoutubeComment{
		AuthorDisplayName:     ct.Snippet.TopLevelComment.Snippet.AuthorDisplayName,
		AuthorProfileImageURL: ct.Snippet.TopLevelComment.Snippet.AuthorProfileImageUrl,
		Text:                  ct.Snippet.TopLevelComment.Snippet.TextOriginal,
		UpdatedAt:             t,
	})

	if ct.Snippet.TotalReplyCount > 0 {
		cs, err := SS.service.Comments.List([]string{"id", "snippet"}).ParentId(ct.Id).Do()
		if err != nil {
			return err
		}

		for _, c := range cs.Items {
			t, _ := time.Parse(time.RFC3339, c.Snippet.UpdatedAt)
			*comments = append(*comments, models.YoutubeComment{
				AuthorDisplayName:     c.Snippet.AuthorDisplayName,
				AuthorProfileImageURL: c.Snippet.AuthorProfileImageUrl,
				Text:                  c.Snippet.TextOriginal,
				ReplyDisplayName:      ct.Snippet.TopLevelComment.Snippet.AuthorDisplayName,
				ReplyText:             ct.Snippet.TopLevelComment.Snippet.TextOriginal,
				UpdatedAt:             t,
			})
		}
	}

	return nil
}

func GetAllComments(videoID string) ([]models.YoutubeComment, error) {
	var err error
	var resp *youtube.CommentThreadListResponse
	comments := []models.YoutubeComment{}

	resp, err = SS.service.CommentThreads.List([]string{"id", "snippet"}).VideoId(videoID).Do()
	if err != nil {
		return comments, err
	}
	for _, ct := range resp.Items {
		err = ProcessCommentThread(ct, &comments)
		if err != nil {
			return comments, err
		}
	}

	for resp.NextPageToken != "" {
		resp, err = SS.service.CommentThreads.List([]string{"id", "snippet"}).VideoId(videoID).PageToken(resp.NextPageToken).Do()
		if err != nil {
			return comments, err
		}
		for _, ct := range resp.Items {
			err = ProcessCommentThread(ct, &comments)
			if err != nil {
				return comments, err
			}
		}
	}

	return comments, nil
}

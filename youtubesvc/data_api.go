package youtubesvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/models"
	"github.com/w8kerr/delubot/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type UserYoutubeService struct {
	service *youtube.Service
	log     *logrus.Entry
}

func (svc *UserYoutubeService) Service() *youtube.Service {
	return svc.service
}

func NewUserYoutubeService(token string, refreshToken *string) (*UserYoutubeService, error) {
	ctx := context.Background()

	log := logrus.WithField("svc", "YoutubeService")
	log.WithField("token", token).WithField("refreshToken", *refreshToken).Info("Initializing user Youtube service")

	// Service account based oauth2 two legged integration
	tokenObj := &oauth2.Token{
		AccessToken: token,
	}
	if refreshToken != nil {
		tokenObj.RefreshToken = *refreshToken
		tokenObj.Expiry = time.Now().Add(-10 * time.Minute)
	}
	log.WithField("token", token).WithField("refreshToken", tokenObj.RefreshToken).Info("Initialized user Youtube service")

	credentialsJSON, err := json.Marshal(config.GoogleOauthCredentials)
	if err != nil {
		log.WithError(err).Error("Failed to initialize service, bad credentials")
		return &UserYoutubeService{}, err
	}
	gConf, err := google.ConfigFromJSON(credentialsJSON, "https://www.googleapis.com/auth/youtube.readonly https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/youtube.force-ssl openid")
	if err != nil {
		log.WithError(err).Error("Failed to initialize service, bad config")
		return &UserYoutubeService{}, err
	}

	// tokenConf := &oauth2.Config{
	// 	ClientID:     config.GoogleClientID,
	// 	ClientSecret: config.GoogleSecret,
	// 	RedirectURL:  "https://localhost:3000/v1/auth/google/callback",
	// 	Endpoint: oauth2.Endpoint{
	// 		AuthURL:   "https://accounts.google.com/o/oauth2/auth",
	// 		TokenURL:  "https://oauth2.googleapis.com/token",
	// 		AuthStyle: oauth2.AuthStyleInParams,
	// 	},
	// 	Scopes: []string{
	// 		"https://www.googleapis.com/auth/youtube.readonly",
	// 		"https://www.googleapis.com/auth/userinfo.profile",
	// 		"https://www.googleapis.com/auth/youtube.force-ssl",
	// 		"openid",
	// 	},
	// }

	client := gConf.Client(ctx, tokenObj)

	service, err := youtube.New(client)
	if err != nil {
		log.WithError(err).Error("Failed to initialize service")
		return &UserYoutubeService{}, err
	}

	return &UserYoutubeService{
		log:     log,
		service: service,
	}, nil
}

func (usvc *UserYoutubeService) SendChatMessage(livechatID string, content string) (*youtube.LiveChatMessage, error) {
	msg := &youtube.LiveChatMessage{
		Snippet: &youtube.LiveChatMessageSnippet{
			LiveChatId: livechatID,
			Type:       "textMessageEvent",
			TextMessageDetails: &youtube.LiveChatTextMessageDetails{
				MessageText: content,
			},
		},
	}
	sent, err := usvc.service.LiveChatMessages.Insert([]string{"snippet"}, msg).Do()
	if err != nil {
		log.Printf("Failed to send chat message: %s", err)
		log.Println(sent, err)
	}
	return sent, err
}

type YoutubeService struct {
	service *youtube.Service
	log     *logrus.Entry
}

func NewYoutubeService(ctx context.Context) (*YoutubeService, error) {
	credentialsJSON, err := json.Marshal(config.GoogleCredentials)
	if err != nil {
		log.Printf("Failed to form Google credentials, %s", err)
		return &YoutubeService{}, err
	}

	// fmt.Println(string(credentialsJSON))

	// Service account based oauth2 two legged integration
	service, err := youtube.NewService(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		log.Printf("Failed to initialize service, %s", err)
		return &YoutubeService{}, err
	}

	return &YoutubeService{
		log:     logrus.WithField("svc", "YoutubeService"),
		service: service,
	}, nil
}

func (svc *YoutubeService) GetStreamInfo(videoID string) (time.Time, *time.Time, *youtube.VideoSnippet, error) {
	resp, err := svc.service.Videos.List([]string{"liveStreamingDetails,snippet"}).Id(videoID).Do()
	if err != nil {
		return time.Time{}, nil, nil, errors.New("Failed to get video info")
	}

	utils.PrintJSON(resp)

	if len(resp.Items) == 0 {
		return time.Time{}, nil, nil, errors.New("Video not found")
	}

	if resp.Items[0].LiveStreamingDetails == nil {
		return time.Time{}, nil, nil, errors.New("Video had no stream details")
	}

	scheduledTimeStr := resp.Items[0].LiveStreamingDetails.ScheduledStartTime
	if scheduledTimeStr == "" {
		return time.Time{}, nil, nil, errors.New("Stream had no start time")
	}
	scheduledTime, err := time.Parse(time.RFC3339, scheduledTimeStr)
	if err != nil {
		return time.Time{}, nil, nil, errors.New("Stream had start time with unexpected format")
	}

	startTimeStr := resp.Items[0].LiveStreamingDetails.ActualStartTime
	if startTimeStr == "" {
		return scheduledTime, nil, resp.Items[0].Snippet, nil
	}
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return scheduledTime, nil, resp.Items[0].Snippet, nil
	}

	return scheduledTime, &startTime, resp.Items[0].Snippet, nil
}

func (svc *YoutubeService) ListUpcomingStreams(channelID string) ([]models.YoutubeStreamRecord, error) {
	liveRecs := []models.YoutubeStreamRecord{}
	resp, err := svc.service.Search.List([]string{"id,snippet"}).ChannelId("UC7YXqPO3eUnxbJ6rN0z2z1Q").Type("video").EventType("upcoming").Do()
	if err != nil {
		return liveRecs, err
	}

	if len(resp.Items) == 0 {
		return liveRecs, nil
	}

	for _, live := range resp.Items {
		vids, err := svc.service.Videos.List([]string{"liveStreamingDetails,snippet"}).Id(live.Id.VideoId).Do()
		if err != nil {
			return liveRecs, err
		}
		vid := vids.Items[0]

		t, _ := time.Parse(time.RFC3339, vid.LiveStreamingDetails.ScheduledStartTime)

		rec := models.YoutubeStreamRecord{
			PostTitle:       vid.Snippet.ChannelTitle,
			PostLink:        "https://www.youtube.com/watch?v=" + vids.Items[0].Id,
			PostPlan:        0,
			YoutubeID:       vids.Items[0].Id,
			Completed:       false,
			ScheduledTime:   t,
			StreamTitle:     vid.Snippet.Title,
			StreamThumbnail: vid.Snippet.Thumbnails.High.Url,
		}

		liveRecs = append(liveRecs, rec)
	}

	return liveRecs, nil
}

func (svc *YoutubeService) GetLivechatID(videoID string) (string, string, error) {
	resp, err := svc.service.Videos.List([]string{"liveStreamingDetails,snippet"}).Id(videoID).Do()
	if err != nil {
		return "", "", errors.New("Failed to get video info")
	}

	if len(resp.Items) == 0 {
		return "", "", errors.New("No video")
	}

	vid := resp.Items[0]
	if vid.LiveStreamingDetails.ActiveLiveChatId == "" {
		return "", vid.Snippet.Title, errors.New("Video is not live")
	}

	fmt.Println("LIVE STREAMING DETAILS")
	utils.PrintJSON(vid)

	return vid.LiveStreamingDetails.ActiveLiveChatId, vid.Snippet.Title, nil
}

var idRE = regexp.MustCompile(`^[^"&?\/\s]{11}$`)
var youtubeRE = regexp.MustCompile(`(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/\s]{11})`)

func (svc *YoutubeService) ParseVideoID(text string) (string, error) {
	idMatches := idRE.FindAllStringSubmatch(text, -1)
	for _, m := range idMatches {
		fmt.Println(m)
		return m[0], nil
	}

	linkMatches := youtubeRE.FindAllStringSubmatch(text, -1)
	for _, m := range linkMatches {
		fmt.Println(m)
		return m[1], nil
	}

	return "", errors.New("No match")
}

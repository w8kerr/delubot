package youtubesvc

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/utils"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

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

	// Service account based oauth2 two legged integration
	service, err := youtube.NewService(ctx, option.WithCredentialsJSON(credentialsJSON))

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

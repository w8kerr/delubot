package tl

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"cloud.google.com/go/translate"
	"github.com/w8kerr/delubot/config"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

var GOOGLE_CLIENT_ID string
var GOOGLE_SECRET string
var GoogleClient *translate.Client

func InitGoogle() {
	credentialsJSON, err := json.Marshal(config.GoogleCredentials)
	if err != nil {
		log.Printf("Failed to form Google credentials, %s", err)
		return
	}

	GOOGLE_CLIENT_ID = os.Getenv("GOOGLE_CLIENT_ID")
	GOOGLE_SECRET = os.Getenv("GOOGLE_SECRET")

	ctx := context.Background()
	GoogleClient, err = translate.NewClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		log.Printf("Failed to create Google Translate client, %s", err)
		return
	}
}

func GoogleTranslate(text string) (string, error) {
	if GoogleClient == nil {
		log.Println("TL Client is not initialized!")
	}

	ctx := context.Background()
	tl, err := GoogleClient.Translate(ctx, []string{text}, language.English, &translate.Options{
		Source: language.Japanese,
		Format: translate.Text,
	})
	if err != nil {
		log.Printf("Failed to translate text '%s', %s", text, err)
		return "", err
	}

	return tl[0].Text, nil
}

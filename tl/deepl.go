package tl

import (
	"log"
	"net/http"
	"os"

	deeplclient "github.com/PineiroHosting/deeplgobindings/pkg"
)

var DEEPL_API_KEY string
var DeepLClient *deeplclient.Client

func InitDeepL() {
	DEEPL_API_KEY = os.Getenv("DEEPL_API_KEY")

	DeepLClient = &deeplclient.Client{
		AuthKey: []byte(DEEPL_API_KEY),
		Client:  &http.Client{},
	}
}

func DeepLTranslate(text string, language deeplclient.ApiLang) (string, string, error) {
	resp, err := DeepLClient.Translate(&deeplclient.TranslationRequest{
		Text:       text,
		TargetLang: language,
	})
	if err != nil {
		log.Printf("Failed to translate text '%s', %s", text, err)
		return "", "", err
	}

	return resp.Translations[0].Text, resp.Translations[0].DetectedSourceLanguage.String(), nil
}

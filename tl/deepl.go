package tl

import (
	"context"
	"log"
	"os"

	"github.com/DaikiYamakawa/deepl-go"
)

var DEEPL_API_KEY string
var DeepLClient *deepl.Client

const (
	// LangDE German
	LangDE = "DE"
	// LangEN English (American)
	LangEN = "EN-US"
	// LangENGB English (British)
	LangENGB = "EN-GB"
	// LangES Spanish
	LangES = "ES"
	// LangFR French
	LangFR = "FR"
	// LangIT Italian
	LangIT = "IT"
	// LangJA Japanese
	LangJA = "JA"
	// LangNL Dutch
	LangNL = "NL"
	// LangPL Polish
	LangPL = "PL"
	// LangPTPT Portuguese (European)
	LangPTPT = "PT-PT"
	// LangPTBR Potuguese (Brazillian)
	LangPTBR = "PT-BR"
	// LangRU Russian
	LangRU = "RU"
	// LangZH Chinese
	LangZH = "ZH"
	// LangAuto Detect
	LangAuto = ""
)

func InitDeepL() {
	DEEPL_API_KEY = os.Getenv("DEEPL_API_KEY")

	client, err := deepl.New("https://api.deepl.com", nil)
	if err != nil {
		log.Printf("Failed to initialize DeepL client: %s", err)
	}

	DeepLClient = client
}

func DeepLTranslate(text string, language string) (string, string, error) {
	resp, err := DeepLClient.TranslateSentence(context.Background(), "Hello", LangAuto, language)
	if err != nil {
		log.Printf("Failed to translate text '%s', %s", text, err)
		return "", "", err
	}

	return resp.Translations[0].Text, resp.Translations[0].DetectedSourceLanguage, nil
}

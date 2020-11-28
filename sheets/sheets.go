package sheets

import (
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

var GOOGLE_CLIENT_ID string
var GOOGLE_SECRET string

func Init(session *discordgo.Session) {
	GOOGLE_CLIENT_ID = os.Getenv("GOOGLE_CLIENT_ID")
	GOOGLE_SECRET = os.Getenv("GOOGLE_SECRET")
}

func Sweeper() {
	sleepDuration := 60 * time.Second
	for {
		time.Sleep(sleepDuration)
		Scan()
	}
}

func Scan() {
	// Get list of users from sheet

	// Process inclusions and exclusions

	// Process big whales

	// roles.EnsureAlphas(verifiedUsers)
	// roles.EnsureWhales()

	fmt.Println("WOULD RUN SCAN")
	fmt.Println(GOOGLE_CLIENT_ID)
	fmt.Println(GOOGLE_SECRET)
}

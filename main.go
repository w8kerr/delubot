// Declare this file to be part of the main package so it can be compiled into
// an executable.
package main

// Import all Go packages required for this file.
import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/sheetsync"
	"github.com/w8kerr/delubot/tl"
	"github.com/w8kerr/delubot/tweetsync"
)

// Version is a constant that stores the Disgord version information.
const Version = "v0.1.0-alpha"

// Session is declared in the global space so it can be easily used
// throughout this program.
// In this use case, there is no error that would be returned.
var Session, _ = discordgo.New()

// Read in all configuration options from both environment variables and
// command line arguments.
func init() {

	// Discord Authentication Token
	Session.Token = os.Getenv("DELUBOT_TOKEN")
	if Session.Token == "" {
		flag.StringVar(&Session.Token, "t", "", "Discord Authentication Token")
	}
}

func main() {

	// Declare any variables needed later.
	var err error

	// Print out a fancy logo!
	fmt.Printf(` 
	________  .__                               .___
	\______ \ |__| ______ ____   ___________  __| _/
	||    |  \|  |/  ___// ___\ /  _ \_  __ \/ __ | 
	||    '   \  |\___ \/ /_/  >  <_> )  | \/ /_/ | 
	||______  /__/____  >___  / \____/|__|  \____ | 
	\_______\/        \/_____/   %-16s\/`+"\n\n", Version)

	// Parse command line arguments
	flag.Parse()

	// Verify a Token was provided
	if Session.Token == "" {
		log.Println("You must provide a Discord authentication token.")
		return
	}

	if !strings.HasPrefix(Session.Token, "Bot ") {
		Session.Token = "Bot " + Session.Token
		fmt.Println("NEW TOKEN")
		fmt.Println(Session.Token)
	}

	// Init mongo
	runtime.GOMAXPROCS(runtime.NumCPU())
	err = mongo.Init(true)
	if err != nil {
		panic("Failed to connect to database")
	}

	config.LoadConfig()

	// Open a websocket connection to Discord
	err = Session.Open()
	if err != nil {
		log.Printf("error opening connection to Discord, %s\n", err)
		os.Exit(1)
	}

	tl.Init()

	env := os.Getenv("DELUBOT_ENV")
	if env != "dev" {
		_, err = Session.ChannelMessageSend("782092598290546722", "🔺DeluBot online!")
		if err != nil {
			fmt.Println(err.Error())
		}

	}

	go tweetsync.InitStreams(Session)

	channels, err := Session.GuildChannels("755437328515989564")
	for _, channel := range channels {
		fmt.Println("Channel viewable - #" + channel.Name)
	}

	sheetsync.Init(Session)
	go sheetsync.Sweeper()

	go Router.InitScanForUpdates(Session)

	// Run the command muxer
	// Session.AddHandler(Router.OnMessageCreate)
	// Router.Route("help", "Display this message.", Router.Help)

	// Wait for a CTRL-C
	log.Printf(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Clean up
	Session.Close()

	// Exit Normally.
}

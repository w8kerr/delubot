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
	"syscall"

	"github.com/bwmarrin/discordgo"

	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/sheets"
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

	// Init mongo
	runtime.GOMAXPROCS(runtime.NumCPU())
	err = mongo.Init(false)
	if err != nil {
		panic("Failed to connect to database")
	}

	// Open a websocket connection to Discord
	err = Session.Open()
	if err != nil {
		log.Printf("error opening connection to Discord, %s\n", err)
		os.Exit(1)
	}

	sheets.Init()
	go sheets.Sweeper()

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

package main

// This file adds the Disgord message route multiplexer, aka "command router".
// to the Disgord bot. This is an optional addition however it is included
// by default to demonstrate how to extend the Disgord bot.

import (
	"fmt"
	"os"

	"github.com/w8kerr/delubot/x/mux"
)

// Router is registered as a global variable to allow easy access to the
// multiplexer throughout the bot.
var Router = mux.New()

func init() {
	// Register the mux OnMessageCreate handler that listens for and processes
	// all messages received.
	Session.AddHandler(Router.OnMessageCreate)

	env := os.Getenv("DELUBOT_ENV")

	// Register the build-in help command.
	if env == "dev" {
		// Dev only commands
	} else {
		// Remote only commands
		Router.Route("help", "Display this message.", Router.Help)
		Router.Route("countmembers", "Count the members on the server.", Router.CountMembers)
		Router.Route("alpharole", "Display or set the configured Alpha role ('clear' to clear).", Router.AlphaRole)
		Router.Route("specialrole", "Display or set the configured Special role ('clear' to clear).", Router.SpecialRole)
		Router.Route("whalerole", "Display or set the configured Whale role ('clear' to clear).", Router.WhaleRole)
		Router.Route("formerrole", "Display or set the configured Former Member role ('clear' to clear).", Router.FormerRole)
		Router.Route("syncsheet", "Display or set the configured Sync Sheet ID ('clear' to clear).", Router.SyncSheet)
		Router.Route("rolegrant", "Check, enable ('enable'), or disable ('disable') role granting.", Router.RoleGrant)
		Router.Route("roleremove", "Check, enable ('enable'), or disable ('disable') role removal.", Router.RoleRemove)
		Router.Route("testsync", "Test what would happen if role syncing was turned on.", Router.TestSync)
		Router.Route("config", "Display all saved configuration objects", Router.Config)
		Router.Route("refreshconfig", "Refresh config from the database", Router.RefreshConfig)
		Router.Route("v", "Close a modmail and copy the verification to the role sync spreadsheet", Router.Verify)
		Router.Route("vd", "Debug the verify command", Router.VDebug)
		Router.Route("addstream", "Add a stream to the schedule manually ('yyyy/mm/dd hh:mm <title>')", Router.AddStream)
		Router.Route("removestream", "Remove a manually added stream ('yyyy/mm/dd hh:mm')", Router.RemoveStream)
		Router.Route("streams", "Display upcoming streams", Router.Streams)
	}
	// Commands for both remote and dev

	fmt.Println("MUX INIT", Router.Prefix)
}

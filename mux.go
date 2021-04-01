package main

// This file adds the Disgord message route multiplexer, aka "command router".
// to the Disgord bot. This is an optional addition however it is included
// by default to demonstrate how to extend the Disgord bot.

import (
	"fmt"
	"os"

	"github.com/w8kerr/delubot/models"
	"github.com/w8kerr/delubot/x/mux"
)

// Router is registered as a global variable to allow easy access to the
// multiplexer throughout the bot.
var Router = mux.New()

func init() {
	// Register the mux OnMessageCreate handler that listens for and processes
	// all messages received.
	Session.AddHandler(Router.OnMessageCreate)
	Session.AddHandler(Router.OnMessageDelete)
	Session.AddHandler(Router.OnMessageDeleteBulk)
	Session.AddHandler(Router.OnMessageUpdate)
	Session.AddHandler(Router.AddReaction)
	Session.AddHandler(Router.RemoveReaction)

	env := os.Getenv("DELUBOT_ENV")

	// Register the build-in help command.
	if env == "dev" {
		// Dev only commands
	} else {
		// Remote only commands
		Router.Route("ytcopy", "Copy messages from the channel to a specified Youtube chat", Router.YoutubeCopy, models.AL_DEV)
		Router.Route("endcopy", "Stop copying messages from the channel to Youtube", Router.EndYoutubeCopy, models.AL_DEV)
		Router.Route("clearuntil", "Clear messages from the channel until reaching the replied-to message", Router.ClearUntil, models.AL_MOD)
		Router.Route("help", "Display this message.", Router.Help, models.AL_STAFF)
		Router.Route("mods", "List people with moderator permissions", Router.Mods, models.AL_MOD)
		Router.Route("countmembers", "Count the members on the server.", Router.CountMembers, models.AL_STAFF)
		Router.Route("alpharole", "Display or set the configured Alpha role ('clear' to clear).", Router.AlphaRole, models.AL_MOD)
		Router.Route("specialrole", "Display or set the configured Special role ('clear' to clear).", Router.SpecialRole, models.AL_MOD)
		Router.Route("whalerole", "Display or set the configured Whale role ('clear' to clear).", Router.WhaleRole, models.AL_MOD)
		Router.Route("formerrole", "Display or set the configured Former Member role ('clear' to clear).", Router.FormerRole, models.AL_MOD)
		Router.Route("muterole", "Display or set the configured Mute role ('clear' to clear).", Router.MuteRole, models.AL_MOD)
		Router.Route("syncsheet", "Display or set the configured Sync Sheet ID ('clear' to clear).", Router.SyncSheet, models.AL_MOD)
		Router.Route("rolegrant", "Check, enable ('enable'), or disable ('disable') role granting.", Router.RoleGrant, models.AL_MOD)
		Router.Route("roleremove", "Check, enable ('enable'), or disable ('disable') role removal.", Router.RoleRemove, models.AL_MOD)
		Router.Route("testsync", "Test what would happen if role syncing was turned on.", Router.TestSync, models.AL_STAFF)
		Router.Route("config", "Display all saved configuration objects", Router.Config, models.AL_MOD)
		Router.Route("refreshconfig", "Refresh config from the database", Router.RefreshConfig, models.AL_STAFF)
		Router.Route("v", "Close a modmail and copy the verification to the role sync spreadsheet", Router.Verify, models.AL_STAFF)
		Router.Route("vd", "Debug the verify command", Router.VDebug, models.AL_STAFF)
		Router.Route("addstream", "Add a stream to the schedule manually ('yyyy/mm/dd hh:mm <title>')", Router.AddStream, models.AL_STAFF)
		Router.Route("removestream", "Remove a manually added stream ('yyyy/mm/dd hh:mm')", Router.RemoveStream, models.AL_STAFF)
		Router.Route("streams", "Display upcoming streams", Router.Streams, models.AL_STAFF)
		Router.Route("stream", "Display upcoming streams", Router.Stream, models.AL_STAFF)
		Router.Route("avatar", "Set the bot avatar", Router.Avatar, models.AL_DEV)
		Router.Route("nickname", "Set the bot nickname", Router.Nickname, models.AL_DEV)
		// Router.Route("proposal", "Create a sign-off sheet following the message.", Router.Proposal, models.AL_STAFF)
		Router.Route("8ball", "Receive the guidance of DeluBot", Router.EightBall, models.AL_STAFF)
		Router.Route("headpat", "Give a headpat", Router.Headpat, models.AL_EVERYONE)
		Router.Route("ttl", "Provide translation for the most recent untranslated tweet in a Twitter feed channel", Router.TweetTranslate, models.AL_STAFF)
		Router.Route("tedit", "Provide translation for the nth tweet (counting upwards) in a Twitter feed channel", Router.TweetEdit, models.AL_STAFF)
		Router.Route("tl", "Translate from Japanese to English", Router.Translate, models.AL_STAFF)
		Router.Route("extractmessages", "Delete an entire segment of chat messages, in between two messages that match a given pattern", Router.ExtractMessages, models.AL_MOD)
		Router.Route("sticky", "Make a message stay at the bottom of the chat", Router.Sticky, models.AL_STAFF)
		Router.Route("unsticky", "Stop promoting the sticky in the current channel", Router.Unsticky, models.AL_STAFF)
	}
	// Commands for both remote and dev

	fmt.Println("MUX INIT", Router.Prefix)
}

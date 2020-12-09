package mux

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (m *Mux) Avatar(ds *discordgo.Session, dm *discordgo.Message, ctx *Context) {
	prerespond := GetResponder(ds, dm)

	ctx.Content = strings.TrimPrefix(ctx.Content, "avatar")
	ctx.Content = strings.TrimSpace(ctx.Content)
	if ctx.Content == "" {
		prerespond("🔺Usage: -db avatar <url>")
		return
	}

	msg := prerespond("🔺Downloading image...")
	respond := GetEditor(ds, msg)

	url := ctx.Content

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		respond("🔺Failed to get image, " + err.Error())
		return
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		respond("🔺Failed to read image, " + err.Error())
	}

	respond("🔺Updating avatar...")
	str := "data:image/png;base64," + base64.StdEncoding.EncodeToString(bytes)
	_, err = ds.UserUpdate("", "", "", str, "")
	if err != nil {
		respond("🔺Failed to update avatar, " + err.Error())
		return
	}

	respond("🔺Avatar updated!")
}

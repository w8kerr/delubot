package config

func Emoji(emojiCode string) string {
	switch emojiCode {
	case "notamusedtea":
		return "<:notamusedtea:774201181425238036>"
	case "delucry":
		return "<:delucry:760035527625539616>"
	case "deluyay":
		return "<:deluyay:759720899737419838>"
	case "delupat":
		return "<:delupat:771927216581771286>"
	case "delucringe":
		return "<:delucringe:789667159076896779>"
	case "white_check_mark":
		return "\u2705"
	case "x":
		return "\u274C"
	default:
		return "[emoji not found]"
	}
}

func NativeEmoji(emojiCode string) string {
	switch emojiCode {
	case "white_check_mark":
		return "\u2705"
	case "x":
		return "\u274C"
	default:
		return "[emoji not found]"
	}
}

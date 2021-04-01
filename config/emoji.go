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
	case "VeePat":
		return "<:VeePat:781367465061515284>"
	case "defaultpat":
		return "<:defaultpat:804958096521953310>"
	case "mirroredpat":
		return "<:mirroredpat:804960286749884448>"
	case "okaytsu":
		return "<:okaytsu:804993887387123732>"
	case "stickpat":
		return "<:stickpat:823128772500652053>"
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

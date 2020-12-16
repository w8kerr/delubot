package config

func Emoji(emojiCode string) string {
	switch emojiCode {
	case "notamusedtea":
		return "<:notamusedtea:774201181425238036>"
	case "delucry":
		return "<:delucry:760035527625539616>"
	case "deluyay":
		return "<:deluyay:759720899737419838>"
	default:
		return "[emoji not found]"
	}
}

package utils

import (
	"golang.org/x/text/message"
)

func HumanNumber(num int) string {
	return message.NewPrinter(message.MatchLanguage("en-US")).Sprintf("%d", num)
}

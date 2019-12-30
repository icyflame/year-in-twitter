package utils

import (
	"math"

	"golang.org/x/text/message"
)

func HumanNumber(num int) string {
	return message.NewPrinter(message.MatchLanguage("en-US")).Sprintf("%d", num)
}

func PercentDiff(num int, bound int, window int) bool {
	return (math.Abs(float64(num-bound)) / float64(bound) * 100) < float64(window)
}

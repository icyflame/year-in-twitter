package tweet

import (
	"time"

	"github.com/dghubble/go-twitter/twitter"
)

type Set []twitter.Tweet

func (s Set) Len() int {
	return len(s)
}

func (s Set) Less(i, j int) bool {
	first_tw_time, _ := time.Parse(time.RubyDate, s[i].CreatedAt)
	second_tw_time, _ := time.Parse(time.RubyDate, s[j].CreatedAt)
	return first_tw_time.Before(second_tw_time)
}

func (s Set) Swap(i, j int) {
	temp := s[i]
	s[i] = s[j]
	s[j] = temp
}

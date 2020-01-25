package tweet

import (
	"sort"
	"testing"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/kylelemons/godebug/pretty"
)

func TestSortTweets(t *testing.T) {
	// Mon Jan 02 15:04:05 -0700 2006
	const date1 = "Thu Jan 23 15:00:00 +0900 2020"
	const date2 = "Fri Jan 24 15:00:00 +0900 2020"
	const date3 = "Sat Jan 25 15:00:00 +0900 2020"

	type testCase struct {
		input      []string
		wantOutput []string
	}

	testCases := map[string]testCase{
		"Base case": {
			input: []string{
				date2, date3, date1,
			},
			wantOutput: []string{
				date1, date2, date3,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			gotTweets := StringsToTweets(tc.input)
			sort.Sort(&gotTweets)
			if diff := pretty.Compare(TweetsToStrings(gotTweets), tc.wantOutput); diff != "" {
				t.Errorf("unexpected result: (-got +want):\n%s", diff)
			}
		})
	}
}

func StringsToTweets(a []string) Set {
	out := make([]twitter.Tweet, len(a))
	for i, s := range a {
		out[i] = twitter.Tweet{
			CreatedAt: s,
		}
	}
	return Set(out)
}

func TweetsToStrings(a []twitter.Tweet) []string {
	out := make([]string, len(a))
	for i, t := range a {
		out[i] = t.CreatedAt
	}
	return out
}

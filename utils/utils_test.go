package utils

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestHumanNumber(t *testing.T) {
	type testCase struct {
		input      int
		wantOutput string
	}

	testCases := map[string]testCase{
		"Base case": {
			input:      4000,
			wantOutput: "4,000",
		},
		"Base case 2": {
			input:      400,
			wantOutput: "400",
		},
		"Base case 3": {
			input:      1234568,
			wantOutput: "1,234,568",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := HumanNumber(tc.input)
			if diff := pretty.Compare(got, tc.wantOutput); diff != "" {
				t.Errorf("Unexpected result: (-got +want):\n%s", diff)
			}
		})
	}
}

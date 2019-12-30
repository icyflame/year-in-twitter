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

func TestPercentDiff(t *testing.T) {
	type testCase struct {
		input       int
		inputBound  int
		inputWindow int
		wantOutput  bool
	}

	testCases := map[string]testCase{
		"Base case": {
			input:       100,
			inputBound:  200,
			inputWindow: 10,
			wantOutput:  false,
		},
		"Base case - true": {
			input:       100,
			inputBound:  110,
			inputWindow: 10,
			wantOutput:  true,
		},
		"Sample case - true": {
			input:       2466,
			inputBound:  3200,
			inputWindow: 50,
			wantOutput:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := PercentDiff(tc.input, tc.inputBound, tc.inputWindow)
			if diff := pretty.Compare(got, tc.wantOutput); diff != "" {
				t.Errorf("Unexpected result: (-got +want):\n%s", diff)
			}
		})
	}
}

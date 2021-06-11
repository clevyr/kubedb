package cli

import (
	"os"
	"strings"
	"testing"
)

func testConfirm(response string, defaultVal bool) error {
	temp := os.Stdout
	os.Stdout = nil
	result := Confirm(strings.NewReader(response + "\n"), defaultVal)
	os.Stdout = temp
	return result
}

func TestConfirm(t *testing.T) {
	type confirmTestCase struct {
		response string
		defaultVal bool
		error error
	}

	testCases := []confirmTestCase{
		{response: "yes", defaultVal: true},
		{response: "yes", defaultVal: false},
		{response: "no", defaultVal: true, error: UserDeclinedErr},
		{response: "no", defaultVal: false, error: UserDeclinedErr},
		{response: "", defaultVal: true},
		{response: "", defaultVal: false, error: UserDeclinedErr},
	}

	for key, testCase := range testCases {
		err := testConfirm(testCase.response, testCase.defaultVal)
		if err != testCase.error {
			t.Errorf("case %d: got %v; expected %v", key, err, testCase.error)
		}
	}
}

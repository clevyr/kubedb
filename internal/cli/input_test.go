package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	testCases := []struct{
		response string
		defaultVal bool
		err error
	}{
		{"yes", true, nil},
		{"yes", false, nil},
		{"no", true, UserDeclinedErr},
		{"no", false, UserDeclinedErr},
		{"", true, nil},
		{"", false, UserDeclinedErr},
	}

	for _, tc := range testCases {
        tc := tc // capture range variable
		t.Run(fmt.Sprintf("respond %v default %v error %v", tc.response, tc.defaultVal, tc.err), func(t *testing.T) {
            t.Parallel()
			temp := os.Stdout
			os.Stdout = nil
			err := Confirm(strings.NewReader(tc.response + "\n"), tc.defaultVal)
			os.Stdout = temp
			if err != tc.err {
				t.Errorf("got %v; expected %v", err, tc.err)
			}
		})
	}
}

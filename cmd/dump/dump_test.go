package dump

import (
	"fmt"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"strings"
	"testing"
	"time"
)

func TestGenerateFilename(t *testing.T) {
	testCases := []struct {
		namespace string
		filetype  sqlformat.Format
		err       error
	}{
		{"test", sqlformat.Gzip, nil},
		{"another", sqlformat.Plain, nil},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(fmt.Sprintf("%v to %v with error %v", tc.namespace, tc.filetype, tc.err), func(t *testing.T) {
			t.Parallel()
			filename, err := Filename{
				Namespace: tc.namespace,
				Format:    tc.filetype,
				Date:      time.Now(),
			}.Generate()
			if err != tc.err {
				t.Error(err)
			}
			expected := tc.namespace + "_"
			if !strings.HasPrefix(filename, expected) {
				t.Errorf("got %v; expected %#v", filename, expected)
			}
		})
	}
}

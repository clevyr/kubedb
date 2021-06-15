package dump

import (
	"fmt"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"strings"
	"testing"
)

func TestGenerateFilename(t *testing.T) {
	testCases := []struct {
		directory string
		namespace string
		filetype  sqlformat.Format
		err       error
	}{
		{".", "test", sqlformat.Gzip, nil},
		{"/home/test", "another", sqlformat.Plain, nil},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v in %v to %v with error %v", tc.directory, tc.namespace, tc.filetype, tc.err), func(t *testing.T) {
			filename, err := generateFilename(tc.directory, tc.namespace, tc.filetype)
			if err != tc.err {
				t.Error(err)
			}
			expected := tc.directory + "/" + tc.namespace + "-"
			if !strings.HasPrefix(filename, expected) {
				t.Errorf("got %v; expected %v", filename, expected)
			}
		})
	}
}

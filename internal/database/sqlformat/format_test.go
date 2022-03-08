package sqlformat

import (
	"fmt"
	"testing"
)

func TestParseFilename(t *testing.T) {
	testCases := []struct{
		filename string
		expected Format
		err error
	}{
		{"dump.sql.gz", Gzip, nil},
		{"dump.sql", Plain, nil},
		{"dump.dmp", Custom, nil},
		{"dump.png", Unknown, UnknownFormatError},
	}
	for _, tc := range testCases {
        tc := tc // capture range variable
		t.Run(fmt.Sprintf("%v to %v with err %v", tc.filename, tc.expected, tc.err), func(t *testing.T) {
            t.Parallel()
			filetype, err := ParseFilename(tc.filename)
			if err != tc.err {
				t.Error(err)
			}
			if filetype != tc.expected {
				t.Errorf("got %v; expected %v", filetype, tc.expected)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	testCases := []struct{
		format string
		expected Format
		err error
	}{
		{"gzip", Gzip, nil},
		{"gz", Gzip, nil},
		{"g", Gzip, nil},
		{"plain", Plain, nil},
		{"sql", Plain, nil},
		{"p", Plain, nil},
		{"custom", Custom, nil},
		{"c", Custom, nil},
		{"png", Unknown, UnknownFormatError},
	}
	for _, tc := range testCases {
        tc := tc // capture range variable
		t.Run(fmt.Sprintf("%v to %v with error %v", tc.format, tc.expected, tc.err), func(t *testing.T) {
            t.Parallel()
			format, err := ParseFormat(tc.format)
			if err != tc.err {
				t.Error(err)
			}
			if format != tc.expected {
				t.Errorf("got %v; expected %v", format, tc.expected)
			}
		})
	}
}

func TestWriteExtension(t *testing.T) {
	testCases := []struct{
		format Format
		expected string
		err error
	}{
		{Gzip, ".sql.gz", nil},
		{Plain, ".sql", nil},
		{Custom, ".dmp", nil},
		{Unknown, "", UnknownFormatError},
	}
	for _, tc := range testCases {
        tc := tc // capture range variable
		t.Run(fmt.Sprintf("%v to %v with error %v", tc.format, tc.expected, tc.err), func(t *testing.T) {
            t.Parallel()
			ext, err := WriteExtension(tc.format)
			if err != tc.err {
				t.Error(err)
			}
			if ext != tc.expected {
				t.Errorf("got %v; expected %v", ext, tc.expected)
			}
		})
	}
}

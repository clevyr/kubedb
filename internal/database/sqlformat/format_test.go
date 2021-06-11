package sqlformat

import "testing"

func TestParseFilename(t *testing.T) {
	type filenameTestCase struct {
		filename string
		filetype uint8
		error   error
	}

	testCases := []filenameTestCase{
		{filename: "dump.sql.gz", filetype: Gzip},
		{filename: "dump.sql", filetype: Plain},
		{filename: "dump.dmp", filetype: Custom},
		{filename: "dump.png", filetype: Unknown, error: UnknownFormatError},
	}
	for key, testCase := range testCases {
		filetype, err := ParseFilename(testCase.filename)
		if err != testCase.error {
			t.Error(err)
		}
		if filetype != testCase.filetype {
			t.Errorf("case %d: got %v; expected %v", key, filetype, testCase.filetype)
		}
	}
}

func TestParseFormat(t *testing.T) {
	type formatTestCase struct {
		format string
		filetype uint8
		error   error
	}

	testCases := []formatTestCase{
		{format: "gzip", filetype: Gzip},
		{format: "gz", filetype: Gzip},
		{format: "g", filetype: Gzip},
		{format: "plain", filetype: Plain},
		{format: "sql", filetype: Plain},
		{format: "p", filetype: Plain},
		{format: "custom", filetype: Custom},
		{format: "c", filetype: Custom},
		{format: "png", filetype: Unknown, error: UnknownFormatError},
	}
	for key, testCase := range testCases {
		format, err := ParseFormat(testCase.format)
		if err != testCase.error {
			t.Error(err)
		}
		if format != testCase.filetype {
			t.Errorf("case %d: got %v; expected %v", key, format, testCase.filetype)
		}
	}
}

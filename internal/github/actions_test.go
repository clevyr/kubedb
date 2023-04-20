package github

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetOutput(t *testing.T) {
	t.Run("new syntax", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}()

		defer func() {
			_ = os.Unsetenv("GITHUB_OUTPUT")
		}()
		if err := os.Setenv("GITHUB_OUTPUT", f.Name()); !assert.NoError(t, err) {
			return
		}

		if err := SetOutput("test", "passed"); !assert.NoError(t, err) {
			return
		}

		if _, err := f.Seek(0, io.SeekStart); !assert.NoError(t, err) {
			return
		}

		var buf strings.Builder
		if _, err := io.Copy(&buf, f); !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, "test=passed\n", buf.String())
	})

	t.Run("deprecated syntax", func(t *testing.T) {
		var buf strings.Builder
		defer func(file io.Writer) {
			output = file
		}(output)
		output = &buf

		if err := SetOutput("test", "passed"); !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "::set-output name=test::passed", buf.String())
	})
}

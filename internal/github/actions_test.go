package github

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetOutput(t *testing.T) {
	f, err := os.CreateTemp("", "")
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	_ = f.Close()

	defer func() {
		_ = os.Unsetenv("GITHUB_OUTPUT")
	}()
	if err := os.Setenv("GITHUB_OUTPUT", f.Name()); !assert.NoError(t, err) {
		return
	}

	if err := SetOutput("test", "passed"); !assert.NoError(t, err) {
		return
	}

	got, err := os.ReadFile(f.Name())
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "test=passed\n", string(got))
}

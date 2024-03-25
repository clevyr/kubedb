package github

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetOutput(t *testing.T) {
	f, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(f.Name())
	}()
	_ = f.Close()

	t.Setenv("GITHUB_OUTPUT", f.Name())

	require.NoError(t, SetOutput("test", "passed"))

	got, err := os.ReadFile(f.Name())
	require.NoError(t, err)

	assert.Equal(t, "test=passed\n", string(got))
}

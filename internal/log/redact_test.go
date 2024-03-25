package log

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedact_Fire(t *testing.T) {
	t.Parallel()
	type args struct {
		key, value string
	}
	tests := []struct {
		name    string
		r       Redact
		args    args
		want    string
		wantErr require.ErrorAssertionFunc
	}{
		{"safe", Redact("PASS"), args{"username", "USER"}, "USER", require.NoError},
		{"fully redacted", Redact("PASS"), args{"password", "PASS"}, "***", require.NoError},
		{"partially redacted", Redact("PASS"), args{"password", "PASSWORD"}, "***WORD", require.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entry := log.WithField(tt.args.key, tt.args.value)

			err := tt.r.Fire(entry)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, entry.Data[tt.args.key])
		})
	}
}

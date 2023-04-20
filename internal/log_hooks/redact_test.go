package log_hooks

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRedact_Fire(t *testing.T) {
	type args struct {
		key, value string
	}
	tests := []struct {
		name    string
		r       Redact
		args    args
		want    string
		wantErr bool
	}{
		{"safe", Redact("PASS"), args{"username", "USER"}, "USER", false},
		{"fully redacted", Redact("PASS"), args{"password", "PASS"}, "***", false},
		{"partially redacted", Redact("PASS"), args{"password", "PASSWORD"}, "***WORD", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := log.WithField(tt.args.key, tt.args.value)

			err := tt.r.Fire(entry)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, entry.Data[tt.args.key])
		})
	}
}

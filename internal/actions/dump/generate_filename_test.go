package dump

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilename_Generate(t *testing.T) {
	type fields struct {
		Namespace string
		Ext       string
		Date      time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"simple", fields{"test", ".sql.gz", time.Time{}}, "test_0001-01-01_000000.sql.gz", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := Filename{
				Namespace: tt.fields.Namespace,
				Ext:       tt.fields.Ext,
				Date:      tt.fields.Date,
			}
			got := vars.Generate()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHelpFilename(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "clevyr_2022-01-09_094100.sql.gz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HelpFilename()
			assert.Equal(t, tt.want, got)
		})
	}
}

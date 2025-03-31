package dump

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilename_Generate(t *testing.T) {
	type fields struct {
		Database  string
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
		{"no database", fields{"", "test", ".sql.gz", time.Time{}}, "test_0001-01-01_000000.sql.gz", false},
		{
			"with database",
			fields{"postgres", "test", ".sql.gz", time.Time{}},
			"test_postgres_0001-01-01_000000.sql.gz",
			false,
		},
		{
			"database matches namespace",
			fields{"test", "test", ".sql.gz", time.Time{}},
			"test_0001-01-01_000000.sql.gz",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := Filename{
				Database:  tt.fields.Database,
				Namespace: tt.fields.Namespace,
				Ext:       tt.fields.Ext,
				Date:      tt.fields.Date,
			}
			got := vars.Generate()
			assert.Equal(t, tt.want, got)
		})
	}
}

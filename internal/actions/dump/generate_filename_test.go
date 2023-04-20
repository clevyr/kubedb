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
			got, err := vars.Generate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
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

func Test_generate(t *testing.T) {
	type args struct {
		vars Filename
		tmpl string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"simple",
			args{Filename{"test", ".sql.gz", time.Time{}}, FilenameTemplate},
			"test_0001-01-01_000000.sql.gz",
			false,
		},
		{
			"invalid template",
			args{Filename{"test", ".sql.gz", time.Time{}}, "{{"},
			"",
			true,
		},
		{
			"invalid function",
			args{Filename{"test", ".sql.gz", time.Time{}}, "{{ .NotARealFunction }}"},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generate(tt.args.vars, tt.args.tmpl)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

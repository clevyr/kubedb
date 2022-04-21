package dump

import (
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"testing"
	"time"
)

func TestFilename_Ext(t *testing.T) {
	type fields struct {
		Namespace string
		Format    sqlformat.Format
		Date      time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"gzip", fields{"test", sqlformat.Gzip, time.Time{}}, ".sql.gz"},
		{"plain", fields{"test", sqlformat.Plain, time.Time{}}, ".sql"},
		{"custom", fields{"test", sqlformat.Custom, time.Time{}}, ".dmp"},
		{"unknown", fields{"test", sqlformat.Unknown, time.Time{}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := Filename{
				Namespace: tt.fields.Namespace,
				Format:    tt.fields.Format,
				Date:      tt.fields.Date,
			}
			got := vars.Ext()
			if got != tt.want {
				t.Errorf("Ext() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilename_Generate(t *testing.T) {
	type fields struct {
		Namespace string
		Format    sqlformat.Format
		Date      time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"simple", fields{"test", sqlformat.Gzip, time.Time{}}, "test_0001-01-01_000000.sql.gz", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := Filename{
				Namespace: tt.fields.Namespace,
				Format:    tt.fields.Format,
				Date:      tt.fields.Date,
			}
			got, err := vars.Generate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Generate() got = %v, want %v", got, tt.want)
			}
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
			if got := HelpFilename(); got != tt.want {
				t.Errorf("HelpFilename() = %v, want %v", got, tt.want)
			}
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
			args{Filename{"test", sqlformat.Gzip, time.Time{}}, FilenameTemplate},
			"test_0001-01-01_000000.sql.gz",
			false,
		},
		{
			"invalid template",
			args{Filename{"test", sqlformat.Gzip, time.Time{}}, "{{"},
			"",
			true,
		},
		{
			"invalid function",
			args{Filename{"test", sqlformat.Gzip, time.Time{}}, "{{ .NotARealFunction }}"},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generate(tt.args.vars, tt.args.tmpl)
			if (err != nil) != tt.wantErr {
				t.Errorf("generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

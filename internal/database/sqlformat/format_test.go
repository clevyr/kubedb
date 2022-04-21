package sqlformat

import (
	"testing"
)

func TestFormat_Set(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		i       Format
		args    args
		wantErr bool
	}{
		{"gzip", Format(0), args{"gzip"}, false},
		{"gz", Format(0), args{"gz"}, false},
		{"g", Format(0), args{"g"}, false},
		{"plain", Format(0), args{"plain"}, false},
		{"sql", Format(0), args{"sql"}, false},
		{"p", Format(0), args{"p"}, false},
		{"custom", Format(0), args{"custom"}, false},
		{"c", Format(0), args{"c"}, false},
		{"png", Format(0), args{"png"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.i.Set(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormat_Type(t *testing.T) {
	tests := []struct {
		name string
		i    Format
		want string
	}{
		{"simple", Format(0), "string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Type(); got != tt.want {
				t.Errorf("Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFilename(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    Format
		wantErr bool
	}{
		{"sql.gz", args{"test.sql.gz"}, Format(1), false},
		{"sql", args{"test.sql"}, Format(2), false},
		{"dmp", args{"test.dmp"}, Format(3), false},
		{"png", args{"test.png"}, Format(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFilename(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFilename() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	type args struct {
		format string
	}
	tests := []struct {
		name    string
		args    args
		want    Format
		wantErr bool
	}{
		{Gzip.String(), args{Gzip.String()}, Gzip, false},
		{"gz", args{"gz"}, Gzip, false},
		{"g", args{"g"}, Gzip, false},
		{Plain.String(), args{Plain.String()}, Plain, false},
		{"sql", args{"sql"}, Plain, false},
		{"p", args{"p"}, Plain, false},
		{Custom.String(), args{Custom.String()}, Custom, false},
		{"c", args{"c"}, Custom, false},
		{"png", args{"png"}, Unknown, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.args.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteExtension(t *testing.T) {
	type args struct {
		outputFormat Format
	}
	tests := []struct {
		name          string
		args          args
		wantExtension string
		wantErr       bool
	}{
		{"gzip", args{Gzip}, ".sql.gz", false},
		{"plain", args{Plain}, ".sql", false},
		{"custom", args{Custom}, ".dmp", false},
		{"unknown", args{Unknown}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExtension, err := WriteExtension(tt.args.outputFormat)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteExtension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExtension != tt.wantExtension {
				t.Errorf("WriteExtension() gotExtension = %v, want %v", gotExtension, tt.wantExtension)
			}
		})
	}
}

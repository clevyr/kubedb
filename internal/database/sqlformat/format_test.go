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

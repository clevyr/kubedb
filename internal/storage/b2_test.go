package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsB2(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"relative local", args{"test.sql"}, false},
		{"absolute local", args{"/home/test/test.sql"}, false},
		{"gcs bucket", args{"b2://test"}, true},
		{"gcs bucket file", args{"b2://test/test.sql"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsB2(tt.args.path))
		})
	}
}

func TestIsB2Dir(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"relative local", args{"test.sql"}, false},
		{"absolute local", args{"/home/test/test.sql"}, false},
		{"gcs bucket", args{"b2://test"}, true},
		{"gcs bucket file", args{"b2://test/test.sql"}, false},
		{"gcs bucket dir", args{"b2://test/subdir/"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsB2Dir(tt.args.path))
		})
	}
}

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGCS(t *testing.T) {
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
		{"gcs bucket", args{"gs://test"}, true},
		{"gcs bucket file", args{"gs://test/test.sql"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsGCS(tt.args.path))
		})
	}
}

func TestIsGCSDir(t *testing.T) {
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
		{"gcs bucket", args{"gs://test"}, true},
		{"gcs bucket file", args{"gs://test/test.sql"}, false},
		{"gcs bucket dir", args{"gs://test/subdir/"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsGCSDir(tt.args.path))
		})
	}
}

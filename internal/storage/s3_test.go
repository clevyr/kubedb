package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsS3(t *testing.T) {
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
		{"s3 bucket", args{"s3://test"}, true},
		{"s3 bucket file", args{"s3://test/test.sql"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsS3(tt.args.path))
		})
	}
}

func TestIsS3Dir(t *testing.T) {
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
		{"s3 bucket", args{"s3://test"}, true},
		{"s3 bucket file", args{"s3://test/test.sql"}, false},
		{"s3 bucket dir", args{"s3://test/subdir/"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsS3Dir(tt.args.path))
		})
	}
}

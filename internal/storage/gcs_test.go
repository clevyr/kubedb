package storage

import "testing"

func TestIsGCS(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			if got := IsGCS(tt.args.path); got != tt.want {
				t.Errorf("IsGCS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGCSDir(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			if got := IsGCSDir(tt.args.path); got != tt.want {
				t.Errorf("IsGCSDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

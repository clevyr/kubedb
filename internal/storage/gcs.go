package storage

import (
	"context"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
)

const GCSSchema = "gs://"

func IsGCS(path string) bool {
	return strings.HasPrefix(path, GCSSchema)
}

func IsGCSDir(path string) bool {
	if !IsGCS(path) {
		return false
	}
	if strings.HasSuffix(path, "/") {
		return true
	}
	trimmed := strings.TrimPrefix(path, GCSSchema)
	return !strings.Contains(trimmed, "/")
}

func UploadGCS(ctx context.Context, key string) (*storage.Writer, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(key)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	w := client.Bucket(u.Host).Object(u.Path).NewWriter(ctx)
	return w, nil
}

func DownloadGCS(ctx context.Context, key string) (*storage.Reader, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(key)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	r, err := client.Bucket(u.Host).Object(u.Path).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	return r, nil
}

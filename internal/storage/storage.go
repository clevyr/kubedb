package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"time"

	"cloud.google.com/go/storage"
)

type Bucket struct {
	Name string
}

type Object struct {
	Prefix       string
	Name         string
	LastModified time.Time
	Size         int64
}

type Client interface {
	ListBuckets(ctx context.Context) iter.Seq2[*Bucket, error]
	ListObjects(ctx context.Context, key string) iter.Seq2[*Object, error]
	PutObject(ctx context.Context, r io.Reader, key string) error
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
}

func IsCloud(path string) bool {
	return IsS3(path) || IsGCS(path)
}

func IsCloudDir(path string) bool {
	return IsS3Dir(path) || IsGCSDir(path)
}

var ErrUnknownPrefix = errors.New("unknown prefix")

func NewClient(ctx context.Context, path string) (Client, error) {
	switch {
	case IsS3(path):
		return NewS3(ctx)
	case IsGCS(path):
		return NewGCS(ctx, storage.ScopeReadWrite, "")
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownPrefix, path)
	}
}

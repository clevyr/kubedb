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
	Name         string
	IsDir        bool
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
	return IsS3(path) || IsGCS(path) || IsB2(path)
}

func IsCloudDir(path string) bool {
	return IsS3Dir(path) || IsGCSDir(path) || IsB2Dir(path)
}

var ErrUnknownPrefix = errors.New("unknown prefix")

func NewClient(ctx context.Context, path string) (Client, error) {
	switch {
	case IsS3(path):
		return NewS3()
	case IsGCS(path):
		return NewGCS(ctx, storage.ScopeReadWrite, "")
	case IsB2(path):
		return NewB2(ctx)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownPrefix, path)
	}
}

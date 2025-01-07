package storage

import (
	"context"
	"errors"
	"iter"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
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

func newGCSClient(ctx context.Context, scope string) (*storage.Client, error) {
	return storage.NewClient(ctx, option.WithScopes(scope))
}

func ListBucketsGCS(ctx context.Context, projectID string) (iter.Seq2[*storage.BucketAttrs, error], int, error) {
	client, err := newGCSClient(ctx, storage.ScopeReadOnly)
	if err != nil {
		return nil, 0, err
	}

	objects := client.Buckets(ctx, projectID)

	return func(yield func(*storage.BucketAttrs, error) bool) {
		for {
			attrs, err := objects.Next()
			if err != nil && errors.Is(err, iterator.Done) {
				return
			}
			if !yield(attrs, err) {
				return
			}
		}
	}, objects.PageInfo().Remaining(), nil
}

func ListObjectsGCS(ctx context.Context, key string) (iter.Seq2[*storage.ObjectAttrs, error], int, error) {
	client, err := newGCSClient(ctx, storage.ScopeReadOnly)
	if err != nil {
		return nil, 0, err
	}

	u, err := url.Parse(key)
	if err != nil {
		return nil, 0, err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	query := &storage.Query{
		Delimiter:                "/",
		Prefix:                   u.Path,
		Projection:               storage.ProjectionNoACL,
		IncludeFoldersAsPrefixes: true,
	}

	objects := client.Bucket(u.Host).Objects(ctx, query)

	return func(yield func(*storage.ObjectAttrs, error) bool) {
		for {
			attrs, err := objects.Next()
			if err != nil && errors.Is(err, iterator.Done) {
				return
			}
			if !yield(attrs, err) {
				return
			}
		}
	}, objects.PageInfo().Remaining(), nil
}

func UploadGCS(ctx context.Context, key string) (*storage.Writer, error) {
	client, err := newGCSClient(ctx, storage.ScopeReadWrite)
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
	client, err := newGCSClient(ctx, storage.ScopeReadOnly)
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

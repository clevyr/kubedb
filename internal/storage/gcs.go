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

func ListBucketsGCS(ctx context.Context, projectID string) iter.Seq2[*storage.BucketAttrs, error] {
	return func(yield func(*storage.BucketAttrs, error) bool) {
		client, err := newGCSClient(ctx, storage.ScopeReadOnly)
		if err != nil {
			yield(nil, err)
			return
		}

		objects := client.Buckets(ctx, projectID)
		for {
			attrs, err := objects.Next()
			if err != nil && errors.Is(err, iterator.Done) {
				return
			}
			if !yield(attrs, err) {
				return
			}
		}
	}
}

func ListObjectsGCS(ctx context.Context, key string) iter.Seq2[*storage.ObjectAttrs, error] {
	return func(yield func(*storage.ObjectAttrs, error) bool) {
		client, err := newGCSClient(ctx, storage.ScopeReadOnly)
		if err != nil {
			yield(nil, err)
			return
		}

		u, err := url.Parse(key)
		if err != nil {
			yield(nil, err)
			return
		}
		u.Path = strings.TrimLeft(u.Path, "/")

		query := &storage.Query{
			Delimiter:                "/",
			Prefix:                   u.Path,
			Projection:               storage.ProjectionNoACL,
			IncludeFoldersAsPrefixes: true,
		}

		objects := client.Bucket(u.Host).Objects(ctx, query)
		for {
			attrs, err := objects.Next()
			if err != nil && errors.Is(err, iterator.Done) {
				return
			}
			if !yield(attrs, err) {
				return
			}
		}
	}
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

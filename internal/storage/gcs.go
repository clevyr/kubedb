package storage

import (
	"context"
	"errors"
	"io"
	"iter"
	"net/url"
	"os"
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

type GCS struct {
	client    *storage.Client
	projectID string
}

func NewGCS(ctx context.Context, scope, projectID string) (*GCS, error) {
	client, err := storage.NewClient(ctx, option.WithScopes(scope))
	if err != nil {
		return nil, err
	}

	if projectID == "" {
		if val := os.Getenv("GOOGLE_CLOUD_PROJECT"); val != "" {
			projectID = val
		} else if val := os.Getenv("GCLOUD_PROJECT"); val != "" {
			projectID = val
		} else if val := os.Getenv("GCP_PROJECT"); val != "" {
			projectID = val
		}
	}

	return &GCS{
		client:    client,
		projectID: projectID,
	}, nil
}

func (g *GCS) ListBuckets(ctx context.Context) iter.Seq2[*Bucket, error] {
	return func(yield func(*Bucket, error) bool) {
		objects := g.client.Buckets(ctx, g.projectID)

		for {
			attrs, err := objects.Next()
			if err != nil && errors.Is(err, iterator.Done) {
				return
			}
			if !yield(&Bucket{Name: attrs.Name}, err) {
				return
			}
		}
	}
}

func (g *GCS) ListObjects(ctx context.Context, key string) iter.Seq2[*Object, error] {
	return func(yield func(*Object, error) bool) {
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

		objects := g.client.Bucket(u.Host).Objects(ctx, query)

		for {
			attrs, err := objects.Next()
			if err != nil && errors.Is(err, iterator.Done) {
				return
			}
			if !yield(&Object{
				Prefix:       attrs.Prefix,
				Name:         attrs.Name,
				LastModified: attrs.Updated,
				Size:         attrs.Size,
			}, err) {
				return
			}
		}
	}
}

func (g *GCS) PutObject(ctx context.Context, r io.Reader, key string) error {
	u, err := url.Parse(key)
	if err != nil {
		return err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	w := g.client.Bucket(u.Host).Object(u.Path).NewWriter(ctx)
	_, err = io.Copy(w, r)
	return err
}

func (g *GCS) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	u, err := url.Parse(key)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	r, err := g.client.Bucket(u.Host).Object(u.Path).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	return r, nil
}

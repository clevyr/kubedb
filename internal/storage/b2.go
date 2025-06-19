package storage

import (
	"context"
	"io"
	"iter"
	"net/url"
	"os"
	"strings"

	"github.com/Backblaze/blazer/b2"
	"k8s.io/apimachinery/pkg/api/errors"
)

const B2Schema = "b2://"

func IsB2(path string) bool {
	return strings.HasPrefix(path, B2Schema)
}

func IsB2Dir(path string) bool {
	if !IsB2(path) {
		return false
	}
	if strings.HasSuffix(path, "/") {
		return true
	}
	trimmed := strings.TrimPrefix(path, B2Schema)
	return !strings.Contains(trimmed, "/")
}

type B2 struct {
	client *b2.Client
}

const (
	b2KeyIDEnv = "B2_APPLICATION_KEY_ID"
	b2KeyEnv   = "B2_APPLICATION_KEY"
)

func NewB2(ctx context.Context) (*B2, error) {
	id := os.Getenv(b2KeyIDEnv)
	key := os.Getenv(b2KeyEnv)
	if id == "" || key == "" {
		return nil, errors.NewUnauthorized("b2 unauthorized: please set " + b2KeyIDEnv + " and " + b2KeyEnv)
	}

	client, err := b2.NewClient(ctx, id, key)
	if err != nil {
		return nil, err
	}

	return &B2{client: client}, nil
}

func (b *B2) ListBuckets(ctx context.Context) iter.Seq2[*Bucket, error] {
	return func(yield func(*Bucket, error) bool) {
		buckets, err := b.client.ListBuckets(ctx)
		if err != nil {
			yield(nil, err)
			return
		}

		for _, bucket := range buckets {
			if !yield(&Bucket{Name: bucket.Name()}, nil) {
				return
			}
		}
	}
}

func (b *B2) ListObjects(ctx context.Context, key string) iter.Seq2[*Object, error] {
	return func(yield func(*Object, error) bool) {
		u, err := url.Parse(key)
		if err != nil {
			yield(nil, err)
			return
		}
		u.Path = strings.TrimLeft(u.Path, "/")

		bucket, err := b.client.Bucket(ctx, u.Host)
		if err != nil {
			yield(nil, err)
			return
		}

		it := bucket.List(ctx, b2.ListDelimiter("/"), b2.ListPrefix(u.Path))

		for it.Next() {
			if it.Err() != nil {
				yield(nil, it.Err())
				return
			}

			obj := it.Object()
			attrs, err := obj.Attrs(ctx)
			if err != nil {
				yield(nil, err)
				return
			}

			myObj := &Object{
				LastModified: attrs.LastModified,
				Size:         attrs.Size,
			}

			if strings.HasSuffix(attrs.Name, "/") {
				myObj.Prefix = attrs.Name
			} else {
				myObj.Name = attrs.Name
			}

			if !yield(myObj, it.Err()) {
				return
			}
		}
	}
}

func (b *B2) PutObject(ctx context.Context, r io.Reader, key string) error {
	u, err := url.Parse(key)
	if err != nil {
		return err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	bucket, err := b.client.Bucket(ctx, u.Host)
	if err != nil {
		return err
	}

	obj := bucket.Object(u.Path)

	w := obj.NewWriter(ctx)
	defer func() {
		_ = w.Close()
	}()

	_, err = io.Copy(w, r)
	if err != nil {
		_ = w.Close()
		_ = obj.Cancel(ctx)
		return err
	}
	return w.Close()
}

func (b *B2) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	u, err := url.Parse(key)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	bucket, err := b.client.Bucket(ctx, u.Host)
	if err != nil {
		return nil, err
	}

	obj := bucket.Object(u.Path)
	r := obj.NewReader(ctx)
	return r, nil
}

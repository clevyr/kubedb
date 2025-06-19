package storage

import (
	"context"
	"io"
	"iter"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"k8s.io/utils/ptr"
)

const S3Schema = "s3://"

func IsS3(path string) bool {
	return strings.HasPrefix(path, S3Schema)
}

func IsS3Dir(path string) bool {
	if !IsS3(path) {
		return false
	}
	if strings.HasSuffix(path, "/") {
		return true
	}
	trimmed := strings.TrimPrefix(path, S3Schema)
	return !strings.Contains(trimmed, "/")
}

type S3 struct {
	client *s3.Client
}

func NewS3(ctx context.Context) (*S3, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)
	return &S3{client: client}, nil
}

func (s *S3) ListBuckets(ctx context.Context) iter.Seq2[*Bucket, error] {
	return func(yield func(*Bucket, error) bool) {
		input := &s3.ListBucketsInput{}

		for {
			buckets, err := s.client.ListBuckets(ctx, input)
			if err != nil {
				yield(nil, err)
				return
			}

			for _, bucket := range buckets.Buckets {
				if !yield(&Bucket{Name: *bucket.Name}, nil) {
					return
				}
			}

			input.ContinuationToken = buckets.ContinuationToken
			if input.ContinuationToken == nil {
				return
			}
		}
	}
}

func (s *S3) ListObjects(ctx context.Context, key string) iter.Seq2[*Object, error] {
	return func(yield func(*Object, error) bool) {
		u, err := url.Parse(key)
		if err != nil {
			yield(nil, err)
			return
		}
		u.Path = strings.TrimLeft(u.Path, "/")

		input := &s3.ListObjectsV2Input{
			Bucket:    ptr.To(u.Host),
			Delimiter: ptr.To("/"),
			Prefix:    ptr.To(u.Path),
		}

		for {
			objects, err := s.client.ListObjectsV2(ctx, input)
			if err != nil {
				yield(nil, err)
				return
			}

			for _, prefix := range objects.CommonPrefixes {
				if !yield(&Object{Prefix: *prefix.Prefix}, nil) {
					return
				}
			}

			for _, object := range objects.Contents {
				if strings.HasSuffix(*object.Key, "/") {
					continue
				}
				if !yield(&Object{
					Name:         *object.Key,
					LastModified: *object.LastModified,
					Size:         *object.Size,
				}, nil) {
					return
				}
			}

			input.ContinuationToken = objects.NextContinuationToken
			if input.ContinuationToken == nil {
				return
			}
		}
	}
}

func (s *S3) PutObject(ctx context.Context, r io.Reader, key string) error {
	u, err := url.Parse(key)
	if err != nil {
		return err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	_, err = manager.NewUploader(s.client).Upload(ctx, &s3.PutObjectInput{
		Bucket: ptr.To(u.Host),
		Key:    ptr.To(u.Path),
		Body:   r,
	})
	return err
}

func (s *S3) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	u, err := url.Parse(key)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	res, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: ptr.To(u.Host),
		Key:    ptr.To(u.Path),
	})
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

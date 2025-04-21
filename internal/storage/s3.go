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

func initAWS(ctx context.Context) (*s3.Client, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)
	return client, nil
}

func ListBucketsS3(ctx context.Context, input *s3.ListBucketsInput) iter.Seq2[*s3.ListBucketsOutput, error] {
	return func(yield func(*s3.ListBucketsOutput, error) bool) {
		client, err := initAWS(ctx)
		if err != nil {
			yield(nil, err)
			return
		}

		if input == nil {
			input = &s3.ListBucketsInput{}
		}

		for {
			buckets, err := client.ListBuckets(ctx, input)
			if !yield(buckets, err) {
				return
			}

			input.ContinuationToken = buckets.ContinuationToken
			if input.ContinuationToken == nil {
				return
			}
		}
	}
}

func ListObjectsS3(ctx context.Context, key string) iter.Seq2[*s3.ListObjectsV2Output, error] {
	return func(yield func(*s3.ListObjectsV2Output, error) bool) {
		client, err := initAWS(ctx)
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

		input := &s3.ListObjectsV2Input{
			Bucket:    ptr.To(u.Host),
			Delimiter: ptr.To("/"),
			Prefix:    ptr.To(u.Path),
		}

		for {
			objects, err := client.ListObjectsV2(ctx, input)
			if !yield(objects, err) {
				return
			}

			input.ContinuationToken = objects.NextContinuationToken
			if input.ContinuationToken == nil {
				return
			}
		}
	}
}

func UploadS3(ctx context.Context, r io.ReadCloser, key string) error {
	defer func(r io.ReadCloser) {
		_ = r.Close()
	}(r)

	client, err := initAWS(ctx)
	if err != nil {
		return err
	}

	u, err := url.Parse(key)
	if err != nil {
		return err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	_, err = manager.NewUploader(client).Upload(ctx, &s3.PutObjectInput{
		Bucket: ptr.To(u.Host),
		Key:    ptr.To(u.Path),
		Body:   r,
	})
	return err
}

func DownloadS3(ctx context.Context, key string) (io.ReadCloser, error) {
	client, err := initAWS(ctx)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(key)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	res, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: ptr.To(u.Host),
		Key:    ptr.To(u.Path),
	})
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

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

func DownloadS3(ctx context.Context, w *S3DownloadPipe, key string) error {
	defer func() {
		_ = w.w.Close()
	}()

	client, err := initAWS(ctx)
	if err != nil {
		return err
	}

	u, err := url.Parse(key)
	if err != nil {
		return err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	downloader := manager.NewDownloader(client)
	downloader.Concurrency = 1
	_, err = downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: ptr.To(u.Host),
		Key:    ptr.To(u.Path),
	})
	return err
}

type S3DownloadPipe struct {
	r   *io.PipeReader
	w   *io.PipeWriter
	off int64
}

func NewS3DownloadPipe() *S3DownloadPipe {
	r, w := io.Pipe()
	return &S3DownloadPipe{
		r:   r,
		w:   w,
		off: 0,
	}
}

func (s *S3DownloadPipe) Read(p []byte) (int, error) {
	return s.r.Read(p)
}

func (s *S3DownloadPipe) WriteAt(p []byte, off int64) (int, error) {
	if s.off != off {
		return 0, io.EOF
	}

	n, err := s.w.Write(p)
	if err != nil {
		return n, err
	}

	s.off += int64(n)
	return n, nil
}

func (s *S3DownloadPipe) Close() error {
	return s.r.Close()
}

func (s *S3DownloadPipe) CloseWithError(err error) error {
	return s.r.CloseWithError(err)
}

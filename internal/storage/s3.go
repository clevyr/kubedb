package storage

import (
	"context"
	"io"
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

func CreateS3Upload(ctx context.Context, r io.ReadCloser, key string) error {
	defer func(r io.ReadCloser) {
		_ = r.Close()
	}(r)

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	u, err := url.Parse(key)
	if err != nil {
		return err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	uploader := manager.NewUploader(s3.NewFromConfig(awsCfg))

	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: ptr.To(u.Host),
		Key:    ptr.To(u.Path),
		Body:   r,
	})
	return err
}

package storage

import (
	"context"
	"io"
	"iter"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gopkg.in/ini.v1"
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
	client *minio.Client
}

type s3Config struct {
	Endpoint string
	Region   string
	Secure   bool
}

func NewS3() (*S3, error) {
	cfg, err := loadS3Config()
	if err != nil {
		return nil, err
	}

	if cfg.Endpoint == "s3.amazonaws.com" {
		cfg.Region = ""
	}

	opts := &minio.Options{
		Creds: credentials.NewChainCredentials([]credentials.Provider{
			&credentials.EnvAWS{},
			&credentials.FileAWSCredentials{},
			&credentials.IAM{},
		}),
		Secure: cfg.Secure,
		Region: cfg.Region,
	}

	client, err := minio.New(cfg.Endpoint, opts)
	if err != nil {
		return nil, err
	}
	return &S3{client: client}, nil
}

func loadS3Config() (s3Config, error) {
	endpoint := os.Getenv("AWS_ENDPOINT_URL")
	region := os.Getenv("AWS_REGION")

	var cfg s3Config
	var err error

	if endpoint != "" {
		cfg.Endpoint, cfg.Secure, err = parseEndpoint(endpoint)
		if err != nil {
			return s3Config{}, err
		}
	}

	if cfg.Endpoint != "" && region != "" {
		cfg.Region = region
		return cfg, nil
	}

	fileEndpoint, fileRegion := loadConfigFromFile()

	if cfg.Endpoint == "" && fileEndpoint != "" {
		cfg.Endpoint, cfg.Secure, err = parseEndpoint(fileEndpoint)
		if err != nil {
			return s3Config{}, err
		}
	}

	if region == "" {
		region = fileRegion
	}
	cfg.Region = region

	if cfg.Endpoint == "" {
		cfg.Endpoint = "s3.amazonaws.com"
		cfg.Secure = true
	}

	return cfg, nil
}

func loadConfigFromFile() (string, string) {
	configFile := os.Getenv("AWS_CONFIG_FILE")
	if configFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", ""
		}
		configFile = filepath.Join(home, ".aws", "config")
	}

	cfg, err := ini.Load(configFile)
	if err != nil {
		return "", ""
	}

	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "default"
	}

	sectionName := profile
	if profile != "default" {
		sectionName = "profile " + profile
	}

	section, err := cfg.GetSection(sectionName)
	if err != nil {
		return "", ""
	}

	var endpoint string
	if k, err := section.GetKey("endpoint_url"); err == nil {
		endpoint = k.String()
	}

	var region string
	if k, err := section.GetKey("region"); err == nil {
		region = k.String()
	}

	return endpoint, region
}

func parseEndpoint(e string) (string, bool, error) {
	if !strings.Contains(e, "://") {
		e = "https://" + e
	}
	u, err := url.Parse(e)
	if err != nil {
		return "", false, err
	}
	return u.Host, u.Scheme == "https", nil
}

func (s *S3) ListBuckets(ctx context.Context) iter.Seq2[*Bucket, error] {
	return func(yield func(*Bucket, error) bool) {
		buckets, err := s.client.ListBuckets(ctx)
		if err != nil {
			yield(nil, err)
			return
		}

		for _, bucket := range buckets {
			if !yield(&Bucket{Name: bucket.Name}, nil) {
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
		bucket := u.Host
		prefix := u.Path

		opts := minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: false,
		}

		for object := range s.client.ListObjects(ctx, bucket, opts) {
			if object.Err != nil {
				yield(nil, object.Err)
				return
			}

			if strings.HasSuffix(object.Key, "/") {
				if !yield(&Object{
					Name:  object.Key,
					IsDir: true,
				}, nil) {
					return
				}
				continue
			}

			if !yield(&Object{
				Name:         object.Key,
				LastModified: object.LastModified,
				Size:         object.Size,
			}, nil) {
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

	// -1 for object size tells MinIO to automatically use multipart upload
	_, err = s.client.PutObject(ctx, u.Host, u.Path, r, -1, minio.PutObjectOptions{})
	return err
}

func (s *S3) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	u, err := url.Parse(key)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimLeft(u.Path, "/")

	return s.client.GetObject(ctx, u.Host, u.Path, minio.GetObjectOptions{})
}

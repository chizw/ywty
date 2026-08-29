package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Options 对齐 PHP S3Storage 的 options（含 OSS/COS 的别名键）。
type S3Options struct {
	AccessKeyID          string `json:"access_key_id"`
	SecretAccessKey      string `json:"secret_access_key"`
	Endpoint             string `json:"endpoint"`
	Region               string `json:"region"`
	Bucket               string `json:"bucket"`
	UsePathStyleEndpoint bool   `json:"use_path_style_endpoint"`
	Root                 string `json:"root"`
}

// s3FS S3 兼容适配器（AWS S3 / MinIO / OSS / COS / R2 等兼容端点）。
type s3FS struct {
	client *minio.Client
	bucket string
	root   string
	ctx    context.Context
}

// NewS3FromRaw 以 options JSON 构建（键与 PHP S3Storage 一致：
// access_key_id / secret_access_key / endpoint / region / bucket / use_path_style_endpoint）。
// OSS/COS 需配置各自的 S3 兼容端点。
func NewS3FromRaw(raw string) (Filesystem, error) {
	var o S3Options
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &o)
	}
	return newS3From(o)
}

func newS3From(o S3Options) (Filesystem, error) {
	if o.AccessKeyID == "" || o.SecretAccessKey == "" || o.Bucket == "" {
		return nil, errors.New("storage: S3 配置缺少 access_key_id / secret_access_key / bucket")
	}
	endpoint := o.Endpoint
	if endpoint == "" {
		endpoint = "s3.amazonaws.com"
	}
	secure := !strings.HasPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(o.AccessKeyID, o.SecretAccessKey, ""),
		Secure: secure,
		Region: o.Region,
	})
	if err != nil {
		return nil, err
	}
	return &s3FS{client: client, bucket: o.Bucket, root: strings.Trim(o.Root, "/"), ctx: context.Background()}, nil
}

func (s *s3FS) key(path string) string {
	if s.root == "" {
		return path
	}
	return s.root + "/" + strings.TrimPrefix(path, "/")
}

func (s *s3FS) Write(path string, data []byte) error {
	_, err := s.client.PutObject(s.ctx, s.bucket, s.key(path), strings.NewReader(string(data)), int64(len(data)),
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	return err
}

func (s *s3FS) AppendOrCreate(path string, data []byte) error {
	if s.Exists(path) {
		return nil
	}
	return s.Write(path, data)
}

func (s *s3FS) Exists(path string) bool {
	_, err := s.client.StatObject(s.ctx, s.bucket, s.key(path), minio.StatObjectOptions{})
	return err == nil
}

func (s *s3FS) Read(path string) ([]byte, error) {
	obj, err := s.client.GetObject(s.ctx, s.bucket, s.key(path), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = obj.Close() }()
	return io.ReadAll(obj)
}

func (s *s3FS) Delete(path string) error {
	return s.client.RemoveObject(s.ctx, s.bucket, s.key(path), minio.RemoveObjectOptions{})
}

// Move S3 无原子 rename：复制后删除。
func (s *s3FS) Move(from, to string) error {
	src := minio.CopySrcOptions{Bucket: s.bucket, Object: s.key(from)}
	dst := minio.CopyDestOptions{Bucket: s.bucket, Object: s.key(to)}
	if _, err := s.client.CopyObject(s.ctx, dst, src); err != nil {
		return err
	}
	return s.Delete(from)
}

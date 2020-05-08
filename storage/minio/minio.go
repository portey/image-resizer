package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/minio/minio-go/v6"
	"github.com/portey/image-resizer/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	SSL             bool
	BucketName      string
	Location        string
	RootPath        string
}

type Storage struct {
	client     *minio.Client
	bucketName string
	rootPath   string
}

func New(config Config) (*Storage, error) {
	client, err := minio.New(config.Endpoint, config.AccessKeyID, config.SecretAccessKey, config.SSL)
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(config.BucketName)
	if err != nil {
		return nil, err
	}

	if !exists {
		if err := client.MakeBucket(config.BucketName, config.Location); err != nil {
			return nil, err
		}
	}

	return &Storage{
		client:     client,
		bucketName: config.BucketName,
		rootPath:   config.RootPath,
	}, nil
}

func (s *Storage) Read(ctx context.Context, path string) (io.Reader, error) {
	res, err := s.client.GetObjectWithContext(ctx, s.bucketName, s.absolutePath(path), minio.GetObjectOptions{})

	return res, toServiceError(err)
}

func (s *Storage) Upload(ctx context.Context, data io.Reader) (string, error) {
	name := s.generatePath("origin")
	return name, s.doUpload(ctx, name, data)
}

func (s *Storage) UploadResized(ctx context.Context, data io.Reader, width, height int) (string, error) {
	name := s.generatePath(fmt.Sprintf("%d_%d", width, height))
	return name, s.doUpload(ctx, name, data)
}

func (s *Storage) doUpload(ctx context.Context, path string, content io.Reader) error {
	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, content)
	if err != nil {
		return toServiceError(err)
	}

	_, err = s.client.PutObjectWithContext(
		ctx,
		s.bucketName,
		s.absolutePath(path),
		buf,
		n,
		minio.PutObjectOptions{},
	)

	return toServiceError(err)
}

func (s *Storage) absolutePath(relativePath string) string {
	return path.Join(s.rootPath, relativePath)
}

func (s *Storage) generatePath(size string) string {
	return path.Join(time.Now().Format("2006/01/02"), size, uuid.NewV4().String()+".jpeg")
}

func toServiceError(err error) error {
	if err == nil {
		return err
	}

	log.Error(err)

	return errors.Internal
}

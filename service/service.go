package service

import (
	"context"
	"io"
	"time"

	"github.com/portey/image-resizer/model"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	Get(ctx context.Context, id string) (*model.Image, error)
	List(ctx context.Context, limit, offset int) ([]*model.Image, error)
	Save(ctx context.Context, version int, image model.Image) error
}

type Resizer interface {
	Resize(ctx context.Context, data io.Reader, output io.Writer) error
}

type Storage interface {
	Read(ctx context.Context, path string) (io.Reader, error)
	Upload(ctx context.Context, data io.Reader) (string, error)
	UploadResized(ctx context.Context, data io.Reader, width, height int) (string, error)
}

type ImageService struct {
	storage Storage
	resizer Resizer
	repo    Repository
}

func (s *ImageService) Upload(ctx context.Context, upload model.ImageUpload, sizes []model.SizeRequest) (*model.Image, error) {
	originalPath, err := s.storage.Upload(ctx, upload.Content)
	if err != nil {
		return nil, err
	}

	image := &model.Image{
		ID:         uuid.NewV4().String(),
		Path:       originalPath,
		ClientName: upload.Filename,
		MimeType:   upload.MimeType,
		Size:       upload.Size,
		UploadAt:   time.Now(),
		Sizes:      []model.Size{},
		Version:    1,
	}

	image, err = s.doResize(ctx, image, upload.Content, sizes)
	if err != nil {
		return nil, err
	}

	return image, s.repo.Save(ctx, 0, *image)
}

func (s *ImageService) Resize(ctx context.Context, id string, sizes []model.SizeRequest) (*model.Image, error) {
	image, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	reader, err := s.storage.Read(ctx, image.Path)
	if err != nil {
		return nil, err
	}

	image, err = s.doResize(ctx, image, reader, sizes)
	if err != nil {
		return nil, err
	}
	version := image.Version
	image.Version++

	return image, s.repo.Save(ctx, version, *image)
}

func (s *ImageService) List(ctx context.Context, limit, offset int) ([]*model.Image, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *ImageService) doResize(ctx context.Context, image *model.Image, content io.Reader, sizes []model.SizeRequest) (*model.Image, error) {
	for _, size := range sizes {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
			if image.HasResizedSize(size.Width, size.Height) {
				continue
			}

			reader, writer := io.Pipe()
			go func() {
				defer func() {
					err := writer.Close()
					if err != nil {
						log.Error("can't close upload writer", err)
					}
				}()
				if err := s.resizer.Resize(ctx, content, writer); err != nil {
					if closeErr := writer.CloseWithError(err); closeErr != nil {
						log.Error("can't close upload writer after resize", err)
					}
				}
			}()

			resizedPath, err := s.storage.UploadResized(ctx, reader, size.Width, size.Height)
			if err != nil {
				return nil, err
			}

			image.AddSize(resizedPath, size.Width, size.Height)
		}
	}

	return image, nil
}

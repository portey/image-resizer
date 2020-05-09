package service

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/portey/image-resizer/errors"
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
	Resize(ctx context.Context, data io.Reader, output io.Writer, width, height int) error
}

type Storage interface {
	Read(ctx context.Context, path string) (io.Reader, error)
	Upload(ctx context.Context, data io.Reader) (string, error)
	UploadResized(ctx context.Context, data io.Reader, width, height int) (string, error)
}

type ImageService struct {
	storage  Storage
	resizer  Resizer
	repo     Repository
	validate *validator.Validate
}

func New(storage Storage, resizer Resizer, repo Repository) *ImageService {
	validate := validator.New()

	return &ImageService{
		validate: validate,
		storage:  storage,
		resizer:  resizer,
		repo:     repo,
	}
}

func (s *ImageService) Upload(ctx context.Context, upload model.ImageUpload, sizes []model.SizeRequest) (*model.Image, error) {
	if err := s.validateParams(upload, sizes); len(err) > 0 {
		return nil, err
	}
	for _, size := range sizes {
		if err := s.validateParams(size); len(err) > 0 {
			return nil, err
		}
	}

	copyContent, originalContent := copyReader(upload.Content)
	originalPath, err := s.storage.Upload(ctx, copyContent)
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

	image, err = s.doResize(ctx, image, originalContent, sizes)
	if err != nil {
		return nil, err
	}

	return image, s.repo.Save(ctx, 0, *image)
}

func (s *ImageService) Resize(ctx context.Context, id string, sizes []model.SizeRequest) (*model.Image, error) {
	if err := s.validateParams(sizes); len(err) > 0 {
		return nil, err
	}

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
	originalContent := content
	var contentCopy io.Reader
	for _, size := range sizes {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
			if image.HasResizedSize(size.Width, size.Height) {
				continue
			}
			contentCopy, originalContent = copyReader(originalContent)
			reader, writer := io.Pipe()
			go func() {
				defer func() {
					err := writer.Close()
					if err != nil {
						log.Error("can't close upload writer", err)
					}
				}()
				if err := s.resizer.Resize(ctx, contentCopy, writer, size.Width, size.Height); err != nil {
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

func (s *ImageService) validateParams(objs ...interface{}) errors.InvalidParams {
	var paramErrors errors.InvalidParams
	for _, obj := range objs {
		err := s.validate.Struct(obj)
		if err == nil {
			continue
		}

		if vErrors, ok := err.(validator.ValidationErrors); ok {
			for _, vError := range vErrors {
				paramErrors = append(paramErrors, errors.InvalidParam{
					Param:   vError.Field(),
					Message: vError.Tag(),
				})
			}

			continue
		}

		log.Error(err)
		return paramErrors
	}

	return paramErrors
}

func copyReader(in io.Reader) (io.Reader, io.Reader) {
	var buf bytes.Buffer
	cc := io.TeeReader(in, &buf)

	return cc, &buf
}

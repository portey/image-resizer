package service

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/portey/image-resizer/errors"
	"github.com/portey/image-resizer/model"
	"github.com/portey/image-resizer/service/mock"
	"github.com/stretchr/testify/assert"
)

func TestImageService_Upload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	content := "Some content"
	contentResized := "Some resized"

	storage := mock.NewMockStorage(ctrl)
	storage.EXPECT().
		Upload(gomock.Eq(ctx), gomock.Any()).
		DoAndReturn(func(_ context.Context, in io.Reader) (string, error) {
			c, err := ioutil.ReadAll(in)
			assert.NoError(t, err)
			assert.Equal(t, content, string(c))

			return "some/path/test.jpg", nil
		})
	storage.EXPECT().
		UploadResized(gomock.Eq(ctx), gomock.Any(), gomock.Eq(100), gomock.Eq(200)).
		DoAndReturn(func(_ context.Context, in io.Reader, _, _ int) (string, error) {
			c, err := ioutil.ReadAll(in)
			assert.NoError(t, err)
			assert.Equal(t, contentResized, string(c))

			return "some/resized/test.jpg", nil
		})

	resizer := mock.NewMockResizer(ctrl)
	resizer.EXPECT().
		Resize(gomock.Eq(ctx), gomock.Any(), gomock.Any(), gomock.Eq(100), gomock.Eq(200)).
		DoAndReturn(func(_ context.Context, in io.Reader, out io.Writer, width, height int) error {
			c, err := ioutil.ReadAll(in)
			assert.NoError(t, err)
			assert.Equal(t, "Some content", string(c))
			assert.Equal(t, 100, width)
			assert.Equal(t, 200, height)

			_, err = out.Write([]byte(contentResized))
			assert.NoError(t, err)

			return nil
		})

	repo := mock.NewMockRepository(ctrl)
	repo.EXPECT().
		Save(gomock.Eq(ctx), gomock.Eq(0), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ int, i model.Image) error {
			assert.Len(t, i.Sizes, 1)
			assert.Equal(t, "original.png", i.ClientName)
			assert.Equal(t, "image/png", i.MimeType)
			assert.Equal(t, int64(123), i.Size)
			assert.Equal(t, "some/path/test.jpg", i.Path)
			assert.Equal(t, "some/resized/test.jpg", i.Sizes[0].Path)
			assert.Equal(t, 100, i.Sizes[0].Width)
			assert.Equal(t, 200, i.Sizes[0].Height)

			return nil
		})

	srv := New(storage, resizer, repo)
	i, err := srv.Upload(ctx, model.ImageUpload{
		Content:  strings.NewReader(content),
		Filename: "original.png",
		Size:     123,
		MimeType: "image/png",
	}, []model.SizeRequest{{
		Width:  100,
		Height: 200,
	}})
	assert.NoError(t, err)
	assert.NotEmpty(t, i.ID)
}

func Test_Validation(t *testing.T) {
	f := func(obj interface{}, err errors.InvalidParams) {
		srv := New(nil, nil, nil)
		actualErr := srv.validateParams(obj)
		assert.Equal(t, err, actualErr)
	}

	//empty upload
	f(model.ImageUpload{}, errors.InvalidParams{
		{
			Param:   "Content",
			Message: "required",
		},
		{
			Param:   "Filename",
			Message: "required",
		},
		{
			Param:   "Size",
			Message: "required",
		},
		{
			Param:   "MimeType",
			Message: "required",
		},
	})

	//invalid mimetype
	f(model.ImageUpload{
		Content:  strings.NewReader("some content"),
		Filename: "test.txt",
		Size:     123000,
		MimeType: "image/unknown",
	}, errors.InvalidParams{
		{
			Param:   "MimeType",
			Message: "eq=image/jpeg|eq=image/png",
		},
	})

	//fully valid
	f(model.ImageUpload{
		Content:  strings.NewReader("some content"),
		Filename: "test.txt",
		Size:     123000,
		MimeType: "image/png",
	}, nil)
}

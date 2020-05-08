package mongo

import (
	"context"
	"os"
	"testing"
	"time"

	serviceerrors "github.com/portey/image-resizer/error"
	"github.com/portey/image-resizer/model"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	uri      = "mongodb://localhost:27017"
	database = "images"
)

func TestRepository_Ping(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "YES" {
		t.Skip()
	}

	ctx := context.Background()
	repo, err := New(ctx, uri, database)
	assert.NoError(t, err)

	err = repo.Ping()
	assert.NoError(t, err)
}

func TestRepository_Save(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "YES" {
		t.Skip()
	}

	ctx := context.Background()
	repo, err := New(ctx, uri, database)
	assert.NoError(t, err)

	image := model.Image{
		ID:       uuid.NewV4().String(),
		UploadAt: time.Now(),
		Path:     "v1path",
		Sizes: []model.Size{{
			Path:   "v1path",
			Width:  100,
			Height: 100,
		}},
		Version: 1,
	}

	err = repo.Save(ctx, 0, image)
	assert.NoError(t, err)

	// saving new version
	image.Version++
	image.Path = "v2path"
	err = repo.Save(ctx, 1, image)
	assert.NoError(t, err)

	// conflict error: saving a version which already exists
	err = repo.Save(ctx, 1, image)
	assert.Error(t, err)
	assert.Equal(t, serviceerrors.RaceCondition, err)
}

func TestRepository_Get(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "YES" {
		t.Skip()
	}

	ctx := context.Background()
	repo, err := New(ctx, uri, database)
	assert.NoError(t, err)

	image := model.Image{
		ID:         uuid.NewV4().String(),
		Path:       "path",
		ClientName: "client_name",
		MimeType:   "mime",
		Size:       123,
		UploadAt:   time.Now(),
		Sizes: []model.Size{{
			Path:   "v1path",
			Width:  100,
			Height: 100,
		}},
		Version: 1,
	}

	err = repo.Save(ctx, 0, image)
	assert.NoError(t, err)

	res, err := repo.Get(ctx, image.ID)
	assert.NoError(t, err)
	assert.Equal(t, image.Path, res.Path)
	assert.Equal(t, image.ClientName, res.ClientName)
	assert.Equal(t, image.MimeType, res.MimeType)
	assert.Equal(t, image.Size, res.Size)
	assert.Equal(t, image.Version, res.Version)
	assert.Len(t, res.Sizes, 1)
	assert.Equal(t, image.Sizes[0].Path, res.Sizes[0].Path)
	assert.Equal(t, image.Sizes[0].Width, res.Sizes[0].Width)
	assert.Equal(t, image.Sizes[0].Height, res.Sizes[0].Height)

	// not found error
	_, err = repo.Get(ctx, uuid.NewV4().String())
	assert.Error(t, err)
	assert.Equal(t, serviceerrors.NotFound, err)
}

func TestRepository_List(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "YES" {
		t.Skip()
	}

	ctx := context.Background()
	repo, err := New(ctx, uri, database)
	assert.NoError(t, err)

	image := model.Image{
		ID:       uuid.NewV4().String(),
		UploadAt: time.Now(),
		Version:  1,
	}

	_, err = repo.collection.DeleteMany(ctx, bson.D{}, options.Delete())
	assert.NoError(t, err)

	err = repo.Save(ctx, 0, image)
	assert.NoError(t, err)

	res, err := repo.List(ctx, 100, 0)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, image.ID, res[0].ID)

	res, err = repo.List(ctx, 100, 1)
	assert.NoError(t, err)
	assert.Len(t, res, 0)
}

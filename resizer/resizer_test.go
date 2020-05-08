package resizer

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/stretchr/testify/assert"
)

func TestResizer_Resize(t *testing.T) {
	r := New()
	ctx := context.Background()

	file, err := os.Open("./fixtures/image.jpg")
	assert.NoError(t, err)

	fileWriter := bytes.Buffer{}
	err = r.Resize(ctx, file, &fileWriter, 200, 100)
	assert.NoError(t, err)

	img, err := imaging.Decode(bytes.NewReader(fileWriter.Bytes()), imaging.AutoOrientation(true))
	assert.NoError(t, err)
	assert.Equal(t, 200, img.Bounds().Max.X)
	assert.Equal(t, 100, img.Bounds().Max.Y)

	err = imaging.Save(img, "./fixtures/small.jpg")
	assert.NoError(t, err)
}

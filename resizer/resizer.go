package resizer

import (
	"context"
	"io"

	"github.com/disintegration/imaging"
	"github.com/portey/image-resizer/errors"
	log "github.com/sirupsen/logrus"
)

type Resizer struct {
}

func New() *Resizer {
	return &Resizer{}
}

func (r *Resizer) Resize(ctx context.Context, data io.Reader, output io.Writer, width, height int) error {
	img, err := imaging.Decode(data, imaging.AutoOrientation(true))
	if err != nil {
		return toServiceErr(err)
	}

	resized := imaging.Resize(img, width, height, imaging.Lanczos)

	return toServiceErr(imaging.Encode(output, resized, imaging.PNG))
}

func toServiceErr(err error) error {
	if err == nil {
		return err
	}

	log.Error(err)

	return errors.Internal
}

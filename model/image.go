package model

import (
	"io"
	"time"
)

type Image struct {
	ID         string    `json:"id" bson:"_id"`
	Path       string    `json:"path" bson:"path"`
	ClientName string    `json:"clientName" bson:"clientName"`
	MimeType   string    `json:"mimeType" bson:"mimeType"`
	Size       int64     `json:"size" bson:"size"`
	UploadAt   time.Time `json:"uploadAt" bson:"uploadAt"`
	Sizes      []Size    `json:"sizes" bson:"sizes"`
	Version    int       `json:"version" bson:"version"`
}

func (i *Image) HasResizedSize(width int, height int) bool {
	for _, size := range i.Sizes {
		if size.Height == height && size.Width == width {
			return true
		}
	}

	return false
}

func (i *Image) AddSize(path string, width int, height int) {
	i.Sizes = append(i.Sizes, Size{
		Path:   path,
		Width:  width,
		Height: height,
	})
}

type Size struct {
	Path   string `json:"path" bson:"path"`
	Width  int    `json:"width" bson:"width"`
	Height int    `json:"height" bson:"height"`
}

type ImageUpload struct {
	Content  io.Reader
	Filename string
	Size     int64
	MimeType string
}

type SizeRequest struct {
	Width  int
	Height int
}

package resolver

import "github.com/portey/image-resizer/service"

//go:generate go run github.com/99designs/gqlgen

type Resolver struct {
	service *service.ImageService
}

func New(service *service.ImageService) *Resolver {
	return &Resolver{
		service: service,
	}
}

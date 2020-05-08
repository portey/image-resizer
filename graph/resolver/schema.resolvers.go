package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/portey/image-resizer/graph/generated"
	"github.com/portey/image-resizer/graph/model"
	servicemodel "github.com/portey/image-resizer/model"
)

func (r *mutationResolver) UploadImage(ctx context.Context, image graphql.Upload, sizes []*model.SizeInput) (*model.Image, error) {
	upload := servicemodel.ImageUpload{
		Content:  image.File,
		Filename: image.Filename,
		Size:     image.Size,
		MimeType: image.ContentType,
	}
	sz := graphQLSizesToModelSizes(sizes)

	i, err := r.service.Upload(ctx, upload, sz)
	if err != nil {
		return nil, err
	}

	return modelImageToGraphQLImage(i), nil
}

func (r *mutationResolver) ResizeImage(ctx context.Context, imageID string, sizes []*model.SizeInput) (*model.Image, error) {
	sz := graphQLSizesToModelSizes(sizes)

	i, err := r.service.Resize(ctx, imageID, sz)
	if err != nil {
		return nil, err
	}

	return modelImageToGraphQLImage(i), nil
}

func (r *queryResolver) Images(ctx context.Context, limit int, offset int) ([]*model.Image, error) {
	list, err := r.service.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	res := make([]*model.Image, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		res = append(res, modelImageToGraphQLImage(item))
	}

	return res, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

func graphQLSizesToModelSizes(sizes []*model.SizeInput) []servicemodel.SizeRequest {
	res := make([]servicemodel.SizeRequest, 0, len(sizes))
	for _, size := range sizes {
		if size == nil {
			continue
		}
		res = append(res, servicemodel.SizeRequest{
			Width:  size.Width,
			Height: size.Height,
		})
	}

	return res
}

func modelImageToGraphQLImage(image *servicemodel.Image) *model.Image {
	if image == nil {
		return nil
	}

	sizes := make([]*model.Size, len(image.Sizes))
	for i, size := range image.Sizes {
		sizes[i] = &model.Size{
			Path:   size.Path,
			Width:  size.Width,
			Height: size.Height,
		}
	}

	return &model.Image{
		ID:         image.ID,
		Path:       image.Path,
		ClientName: image.ClientName,
		MimeType:   image.MimeType,
		Size:       int(image.Size),
		UploadAt:   &image.UploadAt,
		Sizes:      sizes,
	}
}

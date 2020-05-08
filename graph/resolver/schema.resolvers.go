package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/portey/image-resizer/graph/generated"
	"github.com/portey/image-resizer/graph/model"
)

func (r *mutationResolver) UploadImage(ctx context.Context, image graphql.Upload, sizes []*model.SizeInput) (*model.Image, error) {
	return nil, nil
}

func (r *mutationResolver) ResizeImage(ctx context.Context, imageID string, sizes []*model.SizeInput) (*model.Image, error) {
	return nil, nil
}

func (r *queryResolver) Images(ctx context.Context, limit int, offset int) ([]*model.Image, error) {
	return nil, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

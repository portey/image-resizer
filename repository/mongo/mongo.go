package mongo

import (
	"context"
	"time"

	serviceerrors "github.com/portey/image-resizer/error"
	"github.com/portey/image-resizer/model"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collection = "images"

type Repository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func New(ctx context.Context, uri, database string) (*Repository, error) {
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		closeCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if closeErr := client.Disconnect(closeCtx); closeErr != nil {
			log.Error(err)
		}
	}()

	return &Repository{
		client:     client,
		collection: client.Database(database).Collection(collection),
	}, nil
}

func (r *Repository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return r.client.Ping(ctx, nil)
}

func (r *Repository) Get(ctx context.Context, id string) (*model.Image, error) {
	res := r.collection.FindOne(ctx, bson.D{{"_id", id}})
	if res.Err() != nil {
		return nil, toServiceError(res.Err())
	}

	var i model.Image
	if err := res.Decode(&i); err != nil {
		return nil, toServiceError(err)
	}

	return &i, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]*model.Image, error) {
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))

	cur, err := r.collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, toServiceError(err)
	}

	var elems []*model.Image
	for cur.Next(ctx) {
		var elem model.Image
		err := cur.Decode(&elem)
		if err != nil {
			return nil, toServiceError(err)
		}
		elems = append(elems, &elem)
	}

	if err := cur.Err(); err != nil {
		return nil, toServiceError(err)
	}

	if err := cur.Close(ctx); err != nil {
		return nil, toServiceError(err)
	}

	return elems, nil
}

func (r *Repository) Save(ctx context.Context, version int, image model.Image) error {
	filter := bson.D{
		{"_id", image.ID},
		{"version", version},
	}

	if version == 0 {
		_, err := r.collection.InsertOne(ctx, image)
		return toServiceError(err)
	}

	updateResult, err := r.collection.UpdateOne(ctx, filter, bson.D{{"$set", image}}, options.Update())
	if err != nil {
		return toServiceError(err)
	}

	if updateResult.ModifiedCount == 0 {
		return serviceerrors.RaceCondition
	}

	return nil
}

func toServiceError(err error) error {
	if err == nil {
		return nil
	}

	if err == mongo.ErrNoDocuments {
		return serviceerrors.NotFound
	}

	log.Error(err)

	return serviceerrors.Internal
}

package repositories

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"pulse-backend/models"
)

type MongoPollRepository struct {
	collection *mongo.Collection
}

func NewMongoPollRepository(db *mongo.Database) *MongoPollRepository {
	return &MongoPollRepository{
		collection: db.Collection("polls"),
	}
}

func (r *MongoPollRepository) Create(poll *models.Poll) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, poll)

	return err
}

func (r *MongoPollRepository) FindByID(id string) (*models.Poll, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	var poll models.Poll

	err := r.collection.FindOne(
		ctx,
		bson.M{
			"$or": []bson.M{
				{"_id": id},
				{"code": id},
			},
		},
	).Decode(&poll)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("poll not found")
		}

		return nil, err
	}

	return &poll, nil
}

func (r *MongoPollRepository) FindByCreatorID(
	creatorID string,
) ([]*models.Poll, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	cursor, err := r.collection.Find(
		ctx,
		bson.M{
			"creatorId": creatorID,
		},
	)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var polls []*models.Poll

	if err := cursor.All(ctx, &polls); err != nil {
		return nil, err
	}

	return polls, nil
}

func (r *MongoPollRepository) Update(poll *models.Poll) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{
			"_id": poll.ID,
		},
		poll,
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("poll not found")
	}

	return nil
}
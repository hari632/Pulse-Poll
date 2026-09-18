package repositories

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"pulse-backend/models"
)

type MongoVoteRepository struct {
	collection *mongo.Collection
}

func NewMongoVoteRepository(db *mongo.Database) *MongoVoteRepository {
	return &MongoVoteRepository{
		collection: db.Collection("votes"),
	}
}

func (r *MongoVoteRepository) Create(vote *models.Vote) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, vote)

	return err
}

func (r *MongoVoteRepository) CountByPollAndOption(
	pollID string,
	optionID string,
) int {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	count, err := r.collection.CountDocuments(
		ctx,
		bson.M{
			"pollId":   pollID,
			"optionId": optionID,
		},
	)

	if err != nil {
		return 0
	}

	return int(count)
}

func (r *MongoVoteRepository) FindByPollID(
	pollID string,
) ([]*models.Vote, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	cursor, err := r.collection.Find(
		ctx,
		bson.M{
			"pollId": pollID,
		},
	)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var votes []*models.Vote

	if err := cursor.All(ctx, &votes); err != nil {
		return nil, err
	}

	if votes == nil {
		return []*models.Vote{}, nil
	}

	return votes, nil
}

func (r *MongoVoteRepository) FindByVoterAndPoll(
	pollID string,
	voterID string,
) (*models.Vote, error) {
	if voterID == "" {
		return nil, errors.New("voter not found")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	var vote models.Vote

	err := r.collection.FindOne(
		ctx,
		bson.M{
			"pollId":  pollID,
			"voterId": voterID,
		},
	).Decode(&vote)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("vote not found")
		}

		return nil, err
	}

	return &vote, nil
}
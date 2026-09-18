package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDatabase struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func ConnectMongoDB(uri string, databaseName string) (*MongoDatabase, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	client, err := mongo.Connect(
		options.Client().ApplyURI(uri),
	)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(databaseName)

	return &MongoDatabase{
		Client: client,
		DB:     db,
	}, nil
}
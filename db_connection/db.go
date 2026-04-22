package db_connection

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect(ctx context.Context, mongoURI string) (*mongo.Client, *mongo.Database, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, nil, fmt.Errorf("db_connection: failed to connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("db_connection: failed to ping: %w", err)
	}

	db := client.Database("job_search")

	jobs := db.Collection("jobs")
	_, err = jobs.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "userId", Value: 1}}},
		{Keys: bson.D{{Key: "recruiterId", Value: 1}}},
	})
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("db_connection: failed to create jobs indexes: %w", err)
	}

	recruiters := db.Collection("recruiters")
	_, err = recruiters.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "userId", Value: 1}}},
	})
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("db_connection: failed to create recruiters indexes: %w", err)
	}

	return client, db, nil
}

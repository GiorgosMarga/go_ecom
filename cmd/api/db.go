package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func connectDB(cfg config) (*mongo.Database, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(cfg.mongoURI))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client.Database("ecomgo_catbreathe", nil), err
}

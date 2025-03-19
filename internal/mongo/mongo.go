package mongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrUnableToConnect = errors.New("unable to connect to Mongo Client")
	ErrUnableToPing    = errors.New("unable to Ping Mongo Client")
)

type Client struct {
	DB *mongo.Client
}

func NewMongoClient(ctx context.Context, uri string) (*Client, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("NewMongoClient: %w", ErrUnableToConnect)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("NewMongoClient: %w", ErrUnableToPing)
	}

	return &Client{client}, nil
}

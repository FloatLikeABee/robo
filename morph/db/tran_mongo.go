package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TranMongo holds the MongoDB client for Tran (form results stored here).
type TranMongo struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewTranMongo connects to MongoDB and returns the tran database.
func NewTranMongo(uri, database string) (*TranMongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}
	db := client.Database(database)
	return &TranMongo{Client: client, Database: db}, nil
}

// Close disconnects the MongoDB client.
func (m *TranMongo) Close(ctx context.Context) error {
	if m != nil && m.Client != nil {
		return m.Client.Disconnect(ctx)
	}
	return nil
}

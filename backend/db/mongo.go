package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var SalesCollection *mongo.Collection

func ConnectMongo(ctx context.Context) {

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	SalesCollection = client.Database("analyticsDB").Collection("sales")

	fmt.Println("MongoDB Connected")
}

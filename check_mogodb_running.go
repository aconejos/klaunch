package main 

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func check_mongodb_running() error {
	// Replica set addresses
	replSetAddresses := []string{"127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000", "127.0.0.1:27018/?directConnection=true&serverSelectionTimeoutMS=2000", "127.0.0.1:27019/?directConnection=true&serverSelectionTimeoutMS=2000"}

	for _, addr := range replSetAddresses {
			// Attempt to connect to MongoDB
			clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", addr))
			client, _ := mongo.Connect(context.Background(), clientOptions)
			err := client.Ping(context.Background(), nil)
			if err != nil {
				return err
			}
	}	
	fmt.Println("Successfully connected to MongoDB")	
	return nil
}

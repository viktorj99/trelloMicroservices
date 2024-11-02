package repositories

import (
	"context"
	"log"
	"time"
	"users-service/model"

	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection

func InitRepository(client *mongo.Client) {
	userCollection = client.Database("usersDB").Collection("users")
}

func CreateUser(user model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		log.Println("Error inserting user:", err)
		return err
	}
	return nil
}

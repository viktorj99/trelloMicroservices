package repositories

import (
	"context"
	"errors"
	"log"
	"time"
	"users-service/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection

func InitRepository(client *mongo.Client) {
	userCollection = client.Database("usersDB").Collection("users")
}

func CreateUser(user model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existingUser model.User
	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		return errors.New("Email is already in use")
	}
	if err != mongo.ErrNoDocuments {
		log.Println("Error checking existing email:", err)
		return err
	}

	err = userCollection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
	if err == nil {
		return errors.New("Username is already taken")
	}
	if err != mongo.ErrNoDocuments {
		log.Println("Error checking existing username:", err)
		return err
	}

	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		if writeErr, ok := err.(mongo.WriteException); ok {
			for _, e := range writeErr.WriteErrors {
				if e.Code == 11000 {
					return errors.New("Duplicate key error: " + e.Message)
				}
			}
		}
		log.Println("Error inserting user:", err)
		return err
	}
	return nil
}

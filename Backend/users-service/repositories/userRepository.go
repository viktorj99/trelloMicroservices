package repositories

import (
	"context"
	"errors"
	"log"
	"time"
	"users-service/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func GetAllUsers() ([]model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := userCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Error finding users:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			log.Println("Error decoding user:", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		log.Println("Cursor error:", err)
		return nil, err
	}

	return users, nil
}

func GetAllUserMembers() ([]model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"role": "Member"}
	cursor, err := userCollection.Find(ctx, filter)
	if err != nil {
		log.Println("Error finding users:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			log.Println("Error decoding user:", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		log.Println("Cursor error:", err)
		return nil, err
	}

	return users, nil
}

func GetUserByID(userID string) (model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return model.User{}, errors.New("invalid user ID format")
	}

	var user model.User
	err = userCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, errors.New("user not found")
		}
		log.Println("Error finding user by ID:", err)
		return user, err
	}

	return user, nil
}

func GetUserByUsername(username string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user model.User
	err := userCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	} else if err != nil {
		log.Println("Error retrieving user by username:", err)
		return nil, err
	}

	return &user, nil
}

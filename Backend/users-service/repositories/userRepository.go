package repositories

import (
	"context"
	"errors"
	"log"
	"users-service/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var userCollection *mongo.Collection

func InitRepository(client *mongo.Client) {
	userCollection = client.Database("usersDB").Collection("users")
}

func CreateUser(ctx context.Context, user model.User) (interface{}, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "CreateUserRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.email", user.Email),
		attribute.String("user.username", user.Username),
	)

	var existingUser model.User
	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		span.RecordError(errors.New("Email is already in use"))
		return nil, errors.New("Email is already in use")
	}
	if err != mongo.ErrNoDocuments {
		span.RecordError(err)
		log.Println("Error checking existing email:", err)
		return nil, err
	}

	err = userCollection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
	if err == nil {
		span.RecordError(errors.New("Username is already taken"))
		return nil, errors.New("Username is already taken")
	}
	if err != mongo.ErrNoDocuments {
		span.RecordError(err)
		log.Println("Error checking existing username:", err)
		return nil, err
	}

	result, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		span.RecordError(err)
		log.Println("Error inserting user:", err)
		return nil, err
	}

	span.SetAttributes(attribute.String("user.id", result.InsertedID.(primitive.ObjectID).Hex()))
	return result.InsertedID, nil
}

func GetAllUsers(ctx context.Context) ([]model.User, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "GetAllUsersRepo")
	defer span.End()

	cursor, err := userCollection.Find(ctx, bson.M{})
	if err != nil {
		span.RecordError(err)
		log.Println("Error finding users:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			span.RecordError(err)
			log.Println("Error decoding user:", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		span.RecordError(err)
		log.Println("Cursor error:", err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))
	return users, nil
}

func GetAllUserMembers(ctx context.Context) ([]model.User, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "GetAllUserMembersRepo")
	defer span.End()

	filter := bson.M{
		"role":      model.RoleMember,
		"is_active": true,
	}
	span.SetAttributes(attribute.String("filter.role", model.RoleMember))

	cursor, err := userCollection.Find(ctx, filter)
	if err != nil {
		span.RecordError(err)
		log.Println("Error finding users:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			span.RecordError(err)
			log.Println("Error decoding user:", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		span.RecordError(err)
		log.Println("Cursor error:", err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))
	return users, nil
}

func GetUserByID(ctx context.Context, userID string) (model.User, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "GetUserByIDRepo")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID))

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		span.RecordError(err)
		return model.User{}, errors.New("invalid user ID format")
	}

	var user model.User
	err = userCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.RecordError(err)
			return user, errors.New("user not found")
		}
		span.RecordError(err)
		log.Println("Error finding user by ID:", err)
		return user, err
	}

	return user, nil
}

func GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "GetUserByUsernameRepo")
	defer span.End()

	span.SetAttributes(attribute.String("user.username", username))

	var user model.User
	err := userCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		span.RecordError(err)
		return user, errors.New("user not found")
	} else if err != nil {
		span.RecordError(err)
		log.Println("Error retrieving user by username:", err)
		return user, err
	}

	return user, nil
}

func GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "GetUserByEmailRepo")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	var user model.User
	err := userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		span.RecordError(err)
		return user, errors.New("user not found")
	} else if err != nil {
		span.RecordError(err)
		log.Println("Error retrieving user by email:", err)
		return user, err
	}

	return user, nil
}

func UpdateUser(ctx context.Context, userID string, updatedUser model.User) error {
	ctx, span := otel.Tracer("users-service").Start(ctx, "UpdateUserRepo")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID))

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		span.RecordError(err)
		return errors.New("invalid user ID format")
	}

	update := bson.M{
		"$set": bson.M{
			"first_name": updatedUser.FirstName,
			"last_name":  updatedUser.LastName,
			"password":   updatedUser.Password,
			"email":      updatedUser.Email,
			"username":   updatedUser.Username,
			"role":       updatedUser.Role,
			"is_active":  updatedUser.IsActive,
		},
	}

	_, err = userCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		span.RecordError(err)
		log.Println("Error updating user:", err)
		return err
	}

	return nil
}

func DeleteUserByUsername(ctx context.Context, username string) error {
	ctx, span := otel.Tracer("users-service").Start(ctx, "DeleteUserByUsernameRepo")
	defer span.End()

	span.SetAttributes(attribute.String("user.username", username))

	filter := bson.M{"username": username}
	_, err := userCollection.DeleteOne(ctx, filter)
	if err != nil {
		span.RecordError(err)
		log.Printf("Error deleting user by username: %v", err)
		return err
	}

	return nil
}

func DeleteUserById(ctx context.Context, userId string) (*mongo.DeleteResult, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "DeleteUserByIdRepo")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userId))

	objID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		span.RecordError(err)
		return nil, errors.New("invalid user ID format")
	}

	filter := bson.M{"_id": objID}
	result, err := userCollection.DeleteOne(ctx, filter)
	if err != nil {
		span.RecordError(err)
		log.Printf("Error deleting user by ID: %v", err)
		return nil, err
	}

	if result.DeletedCount == 0 {
		err := errors.New("user not found")
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int64("deletedCount", result.DeletedCount))
	return result, nil
}

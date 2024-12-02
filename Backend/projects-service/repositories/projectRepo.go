package repositories

import (
	"context"
	"fmt"
	"log"
	"os"
	"projects-service/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProjectRepo struct {
	cli    *mongo.Client
	logger *log.Logger
	dbName string
}

func NewProjectRepo(ctx context.Context, logger *log.Logger, dbName string) (*ProjectRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")
	if dburi == "" {
		return nil, fmt.Errorf("environment variable MONGO_DB_URI is not set")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		logger.Println("Failed to create MongoDB client:", err)
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		logger.Println("Failed to connect to MongoDB:", err)
		return nil, err
	}

	return &ProjectRepo{
		cli:    client,
		logger: logger,
		dbName: dbName,
	}, nil
}

func (pr *ProjectRepo) Close(ctx context.Context) error {
	return pr.cli.Disconnect(ctx)
}

func (pr *ProjectRepo) collection() *mongo.Collection {
	return pr.cli.Database(pr.dbName).Collection("projects")
}

func (pr *ProjectRepo) GetAll(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	cursor, err := pr.collection().Find(ctx, bson.M{})
	if err != nil {
		pr.logger.Println("Error retrieving all projects:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &projects); err != nil {
		pr.logger.Println("Error decoding all projects:", err)
		return nil, err
	}

	return projects, nil
}

func (pr *ProjectRepo) GetByName(ctx context.Context, name string) (*model.Project, error) {
	var project model.Project
	err := pr.collection().FindOne(ctx, bson.M{"name": name}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		pr.logger.Println("Error retrieving project by name:", err)
		return nil, err
	}

	return &project, nil
}

func (pr *ProjectRepo) GetById(ctx context.Context, id primitive.ObjectID) (*model.Project, error) {
	var project model.Project
	err := pr.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&project)
	if err != nil {
		pr.logger.Println("Error retrieving project by ID:", err)
		return nil, err
	}

	return &project, nil
}

func (pr *ProjectRepo) Insert(ctx context.Context, project *model.Project) (*mongo.InsertOneResult, error) {
	project.ID = primitive.NewObjectID()
	result, err := pr.collection().InsertOne(ctx, project)
	if err != nil {
		pr.logger.Println("Error inserting project:", err)
		return nil, err
	}

	return result, nil
}

func (pr *ProjectRepo) Update(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*mongo.UpdateResult, error) {
	filter := bson.M{"_id": id}

	result, err := pr.collection().UpdateOne(ctx, filter, updateData)
	if err != nil {
		pr.logger.Println("Error updating project:", err)
		return nil, err
	}

	return result, nil
}

func (pr *ProjectRepo) Delete(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	result, err := pr.collection().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		pr.logger.Println("Error deleting project:", err)
		return nil, err
	}

	return result, nil
}

func (pr *ProjectRepo) GetProjectsByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Project, error) {
    fmt.Println("UserID:", userID)
    
    filter := bson.M{"members._id": userID}

    var projects []model.Project
    cursor, err := pr.collection().Find(ctx, filter)
    if err != nil {
        pr.logger.Println("Error retrieving projects by user ID:", err)
        return nil, err
    }
    defer cursor.Close(ctx)

    if err := cursor.All(ctx, &projects); err != nil {
        pr.logger.Println("Error decoding projects by user ID:", err)
        return nil, err
    }

    return projects, nil
}

func (pr *ProjectRepo) GetProjectsByManagerID(ctx context.Context, userID primitive.ObjectID) ([]model.Project, error) {
    fmt.Println("UserID:", userID)
    
    filter := bson.M{"manager._id": userID}

    var projects []model.Project
    cursor, err := pr.collection().Find(ctx, filter)
    if err != nil {
        pr.logger.Println("Error retrieving projects by user ID:", err)
        return nil, err
    }
    defer cursor.Close(ctx)

    if err := cursor.All(ctx, &projects); err != nil {
        pr.logger.Println("Error decoding projects by user ID:", err)
        return nil, err
    }

    return projects, nil
}
package repositories

import (
	"context"
	"log"
	"os"
	"projects-service/model"

	// NoSQL: module containing Mongo api client
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

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &ProjectRepo{
		cli:    client,
		logger: logger,
		dbName: dbName,
	}, nil
}

func (pr *ProjectRepo) collection() *mongo.Collection {
	return pr.cli.Database(pr.dbName).Collection("projects")
}

func (pr *ProjectRepo) GetAll(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	cursor, err := pr.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (pr *ProjectRepo) GetById(ctx context.Context, id primitive.ObjectID) (*model.Project, error) {
	var project model.Project
	err := pr.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&project)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (pr *ProjectRepo) Insert(ctx context.Context, project *model.Project) (*mongo.InsertOneResult, error) {
	project.ID = primitive.NewObjectID()
	result, err := pr.collection().InsertOne(ctx, project)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (pr *ProjectRepo) Update(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*mongo.UpdateResult, error) {
	update := bson.M{"$set": updateData}
	result, err := pr.collection().UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (pr *ProjectRepo) Delete(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	result, err := pr.collection().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}

	return result, nil
}

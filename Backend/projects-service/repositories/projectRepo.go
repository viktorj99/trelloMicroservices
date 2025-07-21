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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	tracer := otel.Tracer("projects-service/repository")
	ctx, span := tracer.Start(ctx, "GetAllProjectsRepo")
	defer span.End()

	span.SetAttributes(attribute.String("db.operation", "find"), attribute.String("db.collection", "projects"))

	var projects []model.Project
	cursor, err := pr.collection().Find(ctx, bson.M{})
	if err != nil {
		pr.logger.Println("Error retrieving all projects:", err)
		span.RecordError(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &projects); err != nil {
		pr.logger.Println("Error decoding all projects:", err)
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("projects.count", len(projects)))
	return projects, nil
}

func (pr *ProjectRepo) GetByName(ctx context.Context, name string) (*model.Project, error) {
	tracer := otel.Tracer("projects-service/repository")
	ctx, span := tracer.Start(ctx, "GetProjectByNameRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "findOne"),
		attribute.String("db.collection", "projects"),
		attribute.String("project.name", name),
	)

	var project model.Project
	err := pr.collection().FindOne(ctx, bson.M{"name": name}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			pr.logger.Println("No project found with name:", name)
			span.SetAttributes(attribute.Bool("db.result.empty", true))
			return nil, nil
		}
		pr.logger.Println("Error retrieving project by name:", err)
		span.RecordError(err)
		return nil, err
	}

	return &project, nil
}

func (pr *ProjectRepo) GetById(ctx context.Context, id primitive.ObjectID) (*model.Project, error) {
	tracer := otel.Tracer("projects-service/repository")
	ctx, span := tracer.Start(ctx, "GetProjectByIdRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "findOne"),
		attribute.String("db.collection", "projects"),
		attribute.String("project.id", id.Hex()),
	)

	var project model.Project
	err := pr.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&project)
	if err != nil {
		span.RecordError(err)
		pr.logger.Println("Error retrieving project by ID:", err)
		if err == mongo.ErrNoDocuments {
			span.SetAttributes(attribute.Bool("db.result.empty", true))
		}
		return nil, err
	}

	return &project, nil
}

func (pr *ProjectRepo) Insert(ctx context.Context, project *model.Project) (*mongo.InsertOneResult, error) {
	tracer := otel.Tracer("projects-service/repository")
	ctx, span := tracer.Start(ctx, "InsertProjectRepo")
	defer span.End()

	project.ID = primitive.NewObjectID()
	span.SetAttributes(
		attribute.String("db.operation", "insert"),
		attribute.String("db.collection", "projects"),
		attribute.String("project.name", project.Name),
	)

	result, err := pr.collection().InsertOne(ctx, project)
	if err != nil {
		span.RecordError(err)
		pr.logger.Println("Error inserting project:", err)
		return nil, err
	}

	span.SetAttributes(attribute.String("db.inserted_id", result.InsertedID.(primitive.ObjectID).Hex()))
	return result, nil
}

func (pr *ProjectRepo) Update(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*mongo.UpdateResult, error) {
	tracer := otel.Tracer("projects-service/repository")
	ctx, span := tracer.Start(ctx, "UpdateProjectRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "update"),
		attribute.String("db.collection", "projects"),
		attribute.String("project.id", id.Hex()),
	)

	filter := bson.M{"_id": id}
	result, err := pr.collection().UpdateOne(ctx, filter, updateData)
	if err != nil {
		span.RecordError(err)
		pr.logger.Println("Error updating project:", err)
		return nil, err
	}

	return result, nil
}

func (pr *ProjectRepo) Delete(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	tracer := otel.Tracer("projects-service/repository")
	ctx, span := tracer.Start(ctx, "DeleteProjectRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "delete"),
		attribute.String("db.collection", "projects"),
		attribute.String("project.id", id.Hex()),
	)

	result, err := pr.collection().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		span.RecordError(err)
		pr.logger.Println("Error deleting project:", err)
		return nil, err
	}

	return result, nil
}

func (pr *ProjectRepo) GetProjectsByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Project, error) {
	tracer := otel.Tracer("projects-service/repository")
	ctx, span := tracer.Start(ctx, "GetProjectsByUserIDRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "find"),
		attribute.String("db.collection", "projects"),
		attribute.String("user.id", userID.Hex()),
	)

	filter := bson.M{"members._id": userID}
	var projects []model.Project
	cursor, err := pr.collection().Find(ctx, filter)
	if err != nil {
		span.RecordError(err)
		pr.logger.Println("Error retrieving projects by user ID:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &projects); err != nil {
		span.RecordError(err)
		pr.logger.Println("Error decoding projects by user ID:", err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("projects.count", len(projects)))
	return projects, nil
}

func (pr *ProjectRepo) GetProjectsByManagerID(ctx context.Context, userID primitive.ObjectID) ([]model.Project, error) {
	tracer := otel.Tracer("projects-service/repository")
	ctx, span := tracer.Start(ctx, "GetProjectsByManagerIDRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "find"),
		attribute.String("db.collection", "projects"),
		attribute.String("manager.id", userID.Hex()),
	)

	filter := bson.M{"manager._id": userID}
	var projects []model.Project
	cursor, err := pr.collection().Find(ctx, filter)
	if err != nil {
		span.RecordError(err)
		pr.logger.Println("Error retrieving projects by manager ID:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &projects); err != nil {
		span.RecordError(err)
		pr.logger.Println("Error decoding projects by manager ID:", err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("projects.count", len(projects)))
	return projects, nil
}

func (pr *ProjectRepo) IsMemberInProject(ctx context.Context, projectId primitive.ObjectID, memberId primitive.ObjectID) (bool, error) {
	tracer := otel.Tracer("projects-service/repository")
	ctx, span := tracer.Start(ctx, "IsMemberInProjectRepo")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "findOne"),
		attribute.String("db.collection", "projects"),
		attribute.String("project.id", projectId.Hex()),
		attribute.String("member.id", memberId.Hex()),
	)

	filter := bson.M{
		"_id":     projectId,
		"members": bson.M{"$elemMatch": bson.M{"_id": memberId}},
	}

	var project model.Project
	err := pr.collection().FindOne(ctx, filter).Decode(&project)
	if err != nil {
		span.RecordError(err)
		if err == mongo.ErrNoDocuments {
			span.SetAttributes(attribute.Bool("db.result.empty", true))
			return false, nil
		}
		pr.logger.Println("Error checking if member is in project:", err)
		return false, err
	}

	return true, nil
}

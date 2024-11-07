package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Status string

const (
	Pending    Status = "PENDING"
	InProgress Status = "IN_PROGRESS"
	Finished   Status = "FINISHED"
)

type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Status      Status             `bson:"status" json:"status"`
	Project     primitive.ObjectID `bson:"project" json:"project"`
	Member      primitive.ObjectID `bson:"member" json:"member"`
	Blocked     bool               `bson:"blocked,omitempty" json:"blocked"`
}

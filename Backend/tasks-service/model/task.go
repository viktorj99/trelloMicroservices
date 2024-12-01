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
	Title       string             `bson:"title" json:"title" validate:"required,min=5,max=100"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Status      Status             `bson:"status" json:"status" validate:"required,oneof=PENDING IN_PROGRESS FINISHED"`
	Project     primitive.ObjectID `bson:"project" json:"project" validate:"required"`
	Member      primitive.ObjectID `bson:"member,omitempty" json:"member"`
	Blocked     bool               `bson:"blocked,omitempty" json:"blocked"`
}

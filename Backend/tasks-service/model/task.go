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
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Status      Status             `bson:"status" json:"status"`
	Project     primitive.ObjectID `bson:"project" json:"project"`
	Member      primitive.ObjectID `bson:"member,omitempty" json:"member"`
	Blocked     bool               `bson:"blocked,omitempty" json:"blocked"`
}

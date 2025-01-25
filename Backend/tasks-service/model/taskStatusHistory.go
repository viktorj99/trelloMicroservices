package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskStatusHistory struct {
   ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
   TaskID    primitive.ObjectID `bson:"task_id" json:"task_id" validate:"required"`
   Status    Status             `bson:"status" json:"status" validate:"required,oneof=PENDING IN_PROGRESS FINISHED"`
   Timestamp time.Time          `bson:"timestamp" json:"timestamp" validate:"required"`
}

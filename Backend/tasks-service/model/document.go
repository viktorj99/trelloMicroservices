package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Document struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TaskID     primitive.ObjectID `bson:"task_id" json:"task_id" validate:"required"`
	FileName   string             `bson:"file_name" json:"file_name" validate:"required"`
	FilePath   string             `bson:"file_path" json:"file_path" validate:"required"`
	UploadedBy primitive.ObjectID `bson:"uploaded_by" json:"uploaded_by" validate:"required"`
	UploadedAt int64              `bson:"uploaded_at" json:"uploaded_at"`
}

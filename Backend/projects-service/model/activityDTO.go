package model

import "time"

type ActivityDTO struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	ProjectID    string    `json:"projectId" bson:"project_id"`
	UserID       string    `json:"userId,omitempty" bson:"user_id"`
	ManagerID    string    `json:"managerId" bson:"manager_id"`
	ActivityType string    `json:"activityType" bson:"activity_type"`
	Description  string    `json:"description,omitempty" bson:"description"`
	Timestamp    time.Time `json:"timestamp" bson:"timestamp"`
}

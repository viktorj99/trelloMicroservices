package model

import "time"

type Activity struct {
	ID           string            `json:"id" bson:"_id,omitempty"`
	ProjectID    string            `json:"projectId" bson:"project_id"`
	TaskID       *string           `json:"taskId,omitempty" bson:"task_id,omitempty"`
	UserID       string            `json:"userId,omitempty" bson:"user_id"`
	ActivityType string            `json:"activityType" bson:"activity_type"`
	Details      map[string]string `json:"details,omitempty" bson:"details"`
	Timestamp    time.Time         `json:"timestamp" bson:"timestamp"`
}

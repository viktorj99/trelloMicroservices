package models

type DependencyRequest struct {
	TaskID      string `json:"taskId"`
	DependentID string `json:"dependentId"`
}

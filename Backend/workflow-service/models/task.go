package models

type Status string

const (
	Pending    Status = "PENDING"
	InProgress Status = "IN_PROGRESS"
	Finished   Status = "FINISHED"
)

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      Status `json:"status"`
	Project     string `json:"project"`
	Member      string `json:"member,omitempty"`
	IsBlocked   bool   `json:"blocked"`
}

type DependencyRequest struct {
	TaskID      string `json:"taskId"`
	DependentID string `json:"dependentId"`
}
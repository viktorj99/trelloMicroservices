package models

type Dependency struct {
	FromTaskID string `json:"fromTaskID"`
	ProjectID  string `json:"projectID"`
	ToTaskID   string `json:"toTaskID"`
}

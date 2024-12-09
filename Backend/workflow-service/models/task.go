package models

type Task struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsBlocked bool   `json:"isBlocked"`
}
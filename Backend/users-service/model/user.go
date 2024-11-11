package model

type User struct {
	ID        string `json:"id,omitempty" bson:"_id,omitempty"`
	FirstName string `json:"first_name" bson:"first_name"`
	LastName  string `json:"last_name" bson:"last_name"`
	Password  string `json:"password" bson:"password"`
	Email     string `json:"email" bson:"email"`
	Username  string `json:"username,omitempty" bson:"username,omitempty"`
	Role      string `json:"role" bson:"role"`
	IsActive  bool   `json:"is_active" bson:"is_active"`
}

const (
	RoleManager = "Manager"
	RoleMember  = "Member"
)

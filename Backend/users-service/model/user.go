package model

type User struct {
	ID        string `json:"id,omitempty" bson:"_id,omitempty"`
	FirstName string `json:"first_name" bson:"first_name" validate:"required,alpha"`
	LastName  string `json:"last_name" bson:"last_name" validate:"required,alpha"`
	Password  string `json:"password" bson:"password" validate:"required,min=6"` // IZMENITI KASNIJE MIN BR KARATKERA ZA LOZ
	Email     string `json:"email" bson:"email" validate:"required,email"`
	Username  string `json:"username,omitempty" bson:"username,omitempty" validate:"required,min=3,max=20"`
	Role      string `json:"role" bson:"role" validate:"required,oneof=Manager Member"`
	IsActive  bool   `json:"is_active" bson:"is_active"`
}

const (
	RoleManager = "Manager"
	RoleMember  = "Member"
)

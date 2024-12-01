package model

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomDate struct {
	time.Time
}

func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if cd.IsZero() {
		return []byte(`null`), nil
	}
	formatted := cd.Format("2006-01-02")
	return json.Marshal(formatted)
}

func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	str := string(data)
	str = strings.Trim(str, `"`)

	parsedTime, err := time.Parse("2006-01-02", str)
	if err == nil {
		cd.Time = parsedTime
		return nil
	}

	parsedTime, err = time.Parse(time.RFC3339, str)
	if err == nil {
		cd.Time = parsedTime
		return nil
	}

	return fmt.Errorf("invalid date format: %s", str)
}

type Project struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name" validate:"required,min=3,max=50"`
	ExpectedEndDate CustomDate         `bson:"expectedEndDate,omitempty" json:"expectedEndDate" validate:"required"`
	MinMembers      int                `bson:"minMembers,omitempty" json:"minMembers" validate:"required,min=1"`
	MaxMembers      int                `bson:"maxMembers,omitempty" json:"maxMembers" validate:"required,gtefield=MinMembers"`
	Manager         User               `bson:"manager,omitempty" json:"manager" validate:"required"`
	Members         []User             `bson:"members,omitempty" json:"members" validate:"required,dive"`
	IsDeleted       bool               `bson:"isDeleted,omitempty" json:"isDeleted"`
}

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username,omitempty" json:"username"`
	Role     string             `bson:"role,omitempty" json:"role"`
}

func (p *Project) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *Project) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(p)
}

func (p *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *User) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(p)
}

package model

import (
	"encoding/json"
	"io"
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
	if str == `null` {
		*cd = CustomDate{}
		return nil
	}
	parsedTime, err := time.Parse(`"2006-01-02"`, str)
	if err != nil {
		return err
	}
	cd.Time = parsedTime
	return nil
}

type Project struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`
	ExpectedEndDate CustomDate         `bson:"expectedEndDate,omitempty" json:"expectedEndDate"`
	MinMembers      int                `bson:"minMembers,omitempty" json:"minMembers"`
	MaxMembers      int                `bson:"maxMembers,omitempty" json:"maxMembers"`
	Manager         User               `bson:"manager,omitempty" json:"manager"`
	Members         []User             `bson:"members,omitempty" json:"members"`
	IsDeleted       bool               `bson:"isDeleted,omitempty" json:"isDeleted"`
}

type User struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Usename string             `bson:"username,omitempty" json:"username"`
	Role    string             `bson:"role,omitempty" json:"role"`
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

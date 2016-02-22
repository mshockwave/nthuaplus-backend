package db

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	Id	bson.ObjectId `bson:"_id,omitempty"`
	Email	string
	Username	string
	FormalId	string
	Thumbnail	string ""

	AuthInfo	UserAuth
}
type UserAuth struct {
	BcryptCost	int
	BcyptHash	string
}

type Reviewer struct {
	Id		bson.ObjectId `bson:"_id,omitempty"`
	BaseProfile	User

	Permissions	[]string
	Topics		[]string
}

type GradeType float64
type RankType	uint32
type ApplicationForm struct {
	Id              bson.ObjectId `bson:"_id,omitempty"`
	OwnerId         bson.ObjectId
	Timestamp       time.Time

			       //Basic Data
	Name            string
	School          string
	Department      string
	SchoolGrade     string
	Birthday        time.Time
	FormalId        string
	Phone           string
	Email           string
	Address         string

			       //Academic Data
	Topic           uint
	Teacher         string
	ResearchArea    string
	ClassHistories  []StudiedClass
	RelatedSkills   string
	AcademicGrade   AcademicGrade
	LangAbilities   []LanguageAbility

			       //Extras
	ResearchPlan    string //File
	Recommendations string //File
	Transcript      string //File
	Others          string //File
}
type StudiedClass struct {
	Name		string
	Semester	string
	Grade		string
}
type AcademicGrade struct {
	Average		GradeType
	Rank		RankType
}
type LanguageAbility struct{
	Name		string

	//Good: 0
	//Average: 1
	//Bad: 2
	Listening	uint
	Speaking	uint
	Reading		uint
	Writing		uint
}
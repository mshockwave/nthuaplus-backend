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
type BasicUser struct {
	Name	string
	Email	string
}

type TopicId	uint
type GradeType float64
type RankType	int32
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
	Topic           TopicId
	Teacher         string
	ResearchArea    string
	ClassHistories  []StudiedClass
	RelatedSkills   string
	AcademicGrade   AcademicGrade
	LangAbilities   []LanguageAbility

			       //Extras
	ResearchPlan    string //File
	Recommendations []string //Recomm entity hashes
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

type BulletinNote struct {
	Id		bson.ObjectId `bson:"_id,omitempty"`

	Title		string ""
	Content		string ""
	TimeStamp	time.Time
}

type Recomm struct {
	Id		bson.ObjectId `bson:"_id,omitempty"`

	Hash		string
	Submitted	bool
	ApplyUser	BasicUser
	Recommender	BasicUser

	Content		string ""
	Attachment	string ""//File
}

type Reviewer struct {
	Id		bson.ObjectId `bson:"_id,omitempty"`
	BaseProfile	User

	Permissions	[]string
	Topics		[]TopicId
}

type ReviewScore uint
type ReviewResult struct {
	Id		bson.ObjectId `bson:"_id,omitempty"`

	Topic		TopicId
	ApplicationId	bson.ObjectId
	ReviewerId	bson.ObjectId

	//Score Data
	ResearchArea	ReviewScore
	Classes		ReviewScore
	Skills		ReviewScore
	Grade		ReviewScore
	Language	ReviewScore
	ResearchPlan	ReviewScore
	Recomm		ReviewScore
	Other		ReviewScore
	Overall		ReviewScore
}
package db

import (
	"gopkg.in/mgo.v2/bson"
	"time"
	"github.com/mshockwave/nthuaplus-backend/public"
)

type User struct {
	Id	bson.ObjectId `bson:"_id,omitempty"`
	Email	string
	Username	string
	FormalId	string
	Thumbnail	string

	AuthInfo	UserAuth

	Permission	public.UserPermission `bson:",omitempty"`
}
type UserAuth struct {
	BcryptCost	int
	BcyptHash	string
}

type GMInfo struct {
	Id	bson.ObjectId `bson:"_id,omitempty"`
	UserId	bson.ObjectId

	//TODO: Permissions
}

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
	Topic           public.TopicId
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
	ApplyUser	public.BasicUser
	Recommender	public.BasicUser

	Content		string ""
	Attachment	string ""//File
}
type RecommEntity struct {
	Id		bson.ObjectId `bson:"_id,omitempty"`
	Hash		string `bson:",omitempty"`

	ApplyUser	public.BasicUser
	Recommender	bson.ObjectId `bson:",omitempty"`

	LastModified	time.Time
	Content		string `bson:",omitempty"`
	Attachment	public.FileStoragePath `bson:",omitempty"`
}

type Reviewer struct {
	Id		bson.ObjectId `bson:"_id,omitempty"`
	BaseProfile	User

	Permissions	[]string
	Topics		[]public.TopicId
}

type ReviewResult struct {
	Id		bson.ObjectId `bson:"_id,omitempty"`

	Topic		public.TopicId
	ApplicationId	bson.ObjectId
	ReviewerId	bson.ObjectId

	//Score Data
	ResearchArea	public.ReviewScore
	Classes		public.ReviewScore
	Skills		public.ReviewScore
	Grade		public.ReviewScore
	Language	public.ReviewScore
	ResearchPlan	public.ReviewScore
	Recomm		public.ReviewScore
	Other		public.ReviewScore
	Overall		public.ReviewScore
}
func (this *ReviewResult) CopyFromReviewResponse(response public.ReviewResponse){
	this.ResearchArea = public.ReviewScore(response.ResearchArea)
	this.Classes = public.ReviewScore(response.Classes)
	this.Skills = public.ReviewScore(response.Skills)
	this.Grade = public.ReviewScore(response.Grade)
	this.Language = public.ReviewScore(response.Language)
	this.ResearchPlan = public.ReviewScore(response.ResearchPlan)
	this.Recomm = public.ReviewScore(response.Recomm)
	this.Other = public.ReviewScore(response.Other)
	this.Overall = public.ReviewScore(response.Overall)
}
package public

import(
	"github.com/mshockwave/nthuaplus-backend/db"
)

type SimpleResult struct {
	Message	string	""
	Description	string	""
}

type UserProfile struct {
	Email	string
	Username	string
	FormalId	string
	Thumbnail	string ""
}

type ReviewerProfile struct {
	Email	string
	Username	string
	FormalId	string
	Thumbnail	string ""

	Topics		[]db.TopicId
	Permissions	[]string
}

type RecommResult  struct {
	Recommender	db.BasicUser
	ApplyUser	db.BasicUser
	Done		bool
	Hash		string "" //Only for reviewers
}

type ReviewResponse struct {
	ResearchArea	int `json:"researchArea"`
	Classes		int `json:"classes"`
	Skills		int `json:"skills"`
	Grade		int `json:"grade"`
	Language	int `json:"language"`
	ResearchPlan	int `json:"researchPlan"`
	Recomm		int `json:"recomm"`
	Other		int `json:"other"`
	Overall		int `json:"overall"`
}
func (this ReviewResponse) CopyToDbReviewResult(result *db.ReviewResult){

	result.ResearchArea = db.ReviewScore(this.ResearchArea)
	result.Classes = db.ReviewScore(this.Classes)
	result.Skills = db.ReviewScore(this.Skills)
	result.Grade = db.ReviewScore(this.Grade)
	result.Language = db.ReviewScore(this.Language)
	result.ResearchPlan = db.ReviewScore(this.ResearchPlan)
	result.Recomm = db.ReviewScore(this.Recomm)
	result.Other = db.ReviewScore(this.Other)
	result.Overall = db.ReviewScore(this.Overall)
}
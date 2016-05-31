package public

import "time"

type SimpleResult struct {
	Message	string	""
	Description	string	""
}

type BasicUser struct {
	Name	string
	Email	string
}

type UserProfile struct {
	Email	string
	Username	string
	FormalId	string
	Thumbnail	string ""
}

type UserPermission uint64
const(
	/*
	Permission Bit mask
	*/
	USER_PERMISSION_NORMAL = 0
	USER_PERMISSION_REVIEW = 1
	USER_PERMISSION_RECOMM = 2

	/*Admin Permission*/
	USER_PERMISSION_GM = 4
)
func (this UserPermission) ContainsPermission(perm_bit uint64) bool {
	this_num := uint64(this)
	return (this_num & perm_bit) != 0
}

type FileStoragePath string

type TopicId	uint
type ReviewerProfile struct {
	Email	string
	Username	string
	FormalId	string
	Thumbnail	string ""

	Topics		[]TopicId
	Permissions	[]string
}

type RecommResult  struct {
	Recommender	BasicUser
	ApplyUser	BasicUser
	Done		bool
	Hash		string "" //Only for reviewers and GMs
}

type RecommView struct {
	Hash		string

	Recommender	BasicUser
	ApplyUser	BasicUser

	LastModified	time.Time
	Content		string
	Attachment	string
}

type ReviewScore uint
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
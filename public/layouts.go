package public

import(
	"../db"
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

type RecommResult  struct {
	Recommender	db.BasicUser
	ApplyUser	db.BasicUser
	Done		bool
}
package public

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

	Topics		[]string
	Permissions	[]string
}

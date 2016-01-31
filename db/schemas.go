package db

import "gopkg.in/mgo.v2/bson"

type User struct {
	Id	bson.ObjectId `bson:"_id,omitempty"`
	Email	string
	Username	string
	FormalId	string

	AuthInfo	UserAuth
}
type UserAuth struct {
	BcryptCost	int
	BcyptHash	string
}

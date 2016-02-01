package db

import "gopkg.in/mgo.v2/bson"

type User struct {
	Id	bson.ObjectId `bson:"_id,omitempty"`
	Email	string
	Username	string
	FormalId	string
	Thumbnail	bson.ObjectId //the image's object id

	AuthInfo	UserAuth
}
type UserAuth struct {
	BcryptCost	int
	BcyptHash	string
}

type Object struct {
	Id	bson.ObjectId `bson:"_id,omitempty"`
	Bucket	string
}

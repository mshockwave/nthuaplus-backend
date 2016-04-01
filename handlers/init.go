package handlers

import "gopkg.in/mgo.v2/bson"

func init(){
	//Init

	exportAppHashMap = make(map[string]bson.ObjectId)
}

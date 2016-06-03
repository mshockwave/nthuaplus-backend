package handlers

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/mshockwave/nthuaplus-backend/public"
	"gopkg.in/mgo.v2/bson"
	"github.com/mshockwave/nthuaplus-backend/db"
	"github.com/mshockwave/nthuaplus-backend/storage"
	"time"
	"strings"
)

func handleRecommDoorBell(resp http.ResponseWriter, req *http.Request){
	_,user_err := public.GetSessionUserId(req)

	if user_err != nil {
		public.ResponseStatusAsJson(resp, 403, nil)
		return
	}

	user_perm,_ := public.GetSessionUserPermission(req)

	if !user_perm.ContainsPermission(public.USER_PERMISSION_RECOMM) {
		public.ResponseStatusAsJson(resp, 403, &public.SimpleResult{
			Message: "Not Recommender",
		})
		return
	}

	public.ResponseOkAsJson(resp, &public.SimpleResult{
		Message: "Ok",
	})
}

func handleViewStagingRecomms(resp http.ResponseWriter, req *http.Request){
	user_id,_ := public.GetSessionUserId(req)

	stagingDb := public.GetNewStagingDatabase()
	defer stagingDb.Session.Close()

	recomms := stagingDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	q := recomms.Find(bson.M{
		"recommender": user_id,
	})
	var result []public.RecommView
	recomm_item := db.RecommEntity{}
	it := q.Iter()
	for it.Next(&recomm_item) {

		signed_url := ""
		if client, err := storage.GetNewStorageClient(); err == nil{
			expireTime := time.Now().Add(time.Duration(1) * time.Hour) //an hour
			signed_url,_ = client.GetNewSignedURL(string(recomm_item.Attachment), expireTime)
		}

		result = append(result, public.RecommView{
			Hash: recomm_item.Hash,
			ApplyUser: recomm_item.ApplyUser,
			Topic: recomm_item.Topic,
			LastModified: recomm_item.LastModified,
			Content: recomm_item.Content,
			Attachment:signed_url,
		})
	}

	public.ResponseOkAsJson(resp, &result)
}


func editStagingRecomm(resp http.ResponseWriter, req *http.Request){

	user_id,_ := public.GetSessionUserId(req)
	vars := mux.Vars(req)
	hash := vars["hash"]

	stagingDb := public.GetNewStagingDatabase()
	defer stagingDb.Session.Close()

	recomm := stagingDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	q := recomm.Find(bson.M{
		"hash": hash,
		"recommender": user_id,
	})
	recomm_result := db.RecommEntity{}
	if e := q.One(&recomm_result); e != nil{
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "No such recomm entry",
		})
		return
	}

	content := req.FormValue("content")
	attachment := req.FormValue("attachment")

	if len(content) > 0{
		recomm_result.Content = content
		recomm_result.LastModified = time.Now()
	}

	if len(attachment) > 0 {
		recomm_result.Attachment = public.FileStoragePath(attachment)
		recomm_result.LastModified = time.Now()
	}

	if e := recomm.UpdateId(recomm_result.Id, &recomm_result); e != nil {
		public.LogE.Printf("Update recomm entity failed: %s\n", e.Error())
		public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
			Message: "Update recomm failed",
		})
	}else{
		public.ResponseOkAsJson(resp, nil)
	}
}
func removeStagingRecomm(resp http.ResponseWriter, req *http.Request){

	user_id,_ := public.GetSessionUserId(req)
	vars := mux.Vars(req)
	hash := vars["hash"]

	stagingDb := public.GetNewStagingDatabase()
	defer stagingDb.Session.Close()

	recomm := stagingDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	q := recomm.Find(bson.M{
		"hash": hash,
		"recommender": user_id,
	})
	recomm_result := db.RecommEntity{}
	if e := q.One(&recomm_result); e != nil{
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "No such recomm entry",
		})
		return
	}

	if e := recomm.RemoveId(recomm_result.Id); e != nil {
		public.LogE.Printf("Remove recomm entity failed: %s\n", e.Error())
		public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
			Message: "Remove recomm failed",
		})
	}else{
		public.ResponseOkAsJson(resp, nil)
	}
}
func viewStagingRecomm(resp http.ResponseWriter, req *http.Request){

	user_perm,_ := public.GetSessionUserPermission(req)
	user_id,_ := public.GetSessionUserId(req)
	vars := mux.Vars(req)
	hash := vars["hash"]

	stagingDb := public.GetNewStagingDatabase()
	defer stagingDb.Session.Close()

	recomm := stagingDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	q := recomm.Find(bson.M{
		"hash": hash,
	})
	recomm_result := db.RecommEntity{}
	if e := q.One(&recomm_result); e != nil{
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "No such recomm entry",
		})
		return
	}

	// Permission check
	if recomm_result.Recommender != user_id && user_perm != public.USER_PERMISSION_GM {
		public.ResponseStatusAsJson(resp, 403, nil)
		return
	}

	signed_url := ""
	if client, err := storage.GetNewStorageClient(); err == nil{
		expireTime := time.Now().Add(time.Duration(1) * time.Hour) //an hour
		signed_url,_ = client.GetNewSignedURL(string(recomm_result.Attachment), expireTime)
	}

	public.ResponseOkAsJson(resp, &public.RecommView{
		Hash: recomm_result.Hash,
		ApplyUser: recomm_result.ApplyUser,
		Topic: recomm_result.Topic,
		LastModified: recomm_result.LastModified,
		Content: recomm_result.Content,
		Attachment:signed_url,
	})
}

func handleInspectStagingRecomm(resp http.ResponseWriter, req *http.Request){
	user_perm,_ := public.GetSessionUserPermission(req)

	if !user_perm.ContainsPermission(public.USER_PERMISSION_RECOMM) {
		public.ResponseStatusAsJson(resp, 403, &public.SimpleResult{
			Message: "Not Recommender",
		})
		return
	}

	switch strings.ToLower(req.Method) {
	case "get":
		viewStagingRecomm(resp, req)
		break

	case "post":
		editStagingRecomm(resp, req)
		break;
	case "put":
		editStagingRecomm(resp, req)
		break

	case "delete":
		removeStagingRecomm(resp, req)
		break

	default:
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "No such http method",
		})
	}
}

func handleRecommFileUpload(resp http.ResponseWriter, req *http.Request){
	user_perm,_ := public.GetSessionUserPermission(req)

	if user_perm.ContainsPermission(public.USER_PERMISSION_RECOMM) {

		handleFormFileUpload(resp, req)

	}else{
		public.ResponseStatusAsJson(resp, 403, &public.SimpleResult{
			Message: "Not Recommender",
		})
	}
}

func handleViewFormalRecomm(resp http.ResponseWriter, req *http.Request){
	user_id,_ := public.GetSessionUserId(req)

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	recomms := appDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	q := recomms.Find(bson.M{
		"recommender": user_id,
	})
	var result []public.RecommView
	recomm_item := db.RecommEntity{}
	it := q.Iter()
	for it.Next(&recomm_item) {

		signed_url := ""
		if client, err := storage.GetNewStorageClient(); err == nil{
			expireTime := time.Now().Add(time.Duration(1) * time.Hour) //an hour
			signed_url,_ = client.GetNewSignedURL(string(recomm_item.Attachment), expireTime)
		}

		result = append(result, public.RecommView{
			Hash: recomm_item.Hash,
			ApplyUser: recomm_item.ApplyUser,
			Topic: recomm_item.Topic,
			LastModified: recomm_item.LastModified,
			Content: recomm_item.Content,
			Attachment:signed_url,
		})
	}

	public.ResponseOkAsJson(resp, &result)
}

func handleViewRecomm(resp http.ResponseWriter, req *http.Request){
	user_id,_ :=  public.GetSessionUserId(req)
	user_perm,_ := public.GetSessionUserPermission(req)

	vars := mux.Vars(req)
	hash_str := vars["hash"]

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	recomm := appDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	q := recomm.Find(bson.M{
		"hash": hash_str,
	})

	recomm_result := db.RecommEntity{}
	if e := q.One(&recomm_result); e != nil{
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "No such recomm entity",
		})
		return
	}

	// Permission check
	if recomm_result.Recommender != user_id && user_perm != public.USER_PERMISSION_GM {
		public.ResponseStatusAsJson(resp, 403, nil)
		return
	}

	signed_url := ""
	if client, err := storage.GetNewStorageClient(); err == nil{
		expireTime := time.Now().Add(time.Duration(1) * time.Hour) //an hour
		signed_url,_ = client.GetNewSignedURL(string(recomm_result.Attachment), expireTime)
	}

	public.ResponseOkAsJson(resp, &public.RecommView{
		Hash: recomm_result.Hash,
		ApplyUser: recomm_result.ApplyUser,
		Topic: recomm_result.Topic,
		LastModified: recomm_result.LastModified,
		Content: recomm_result.Content,
		Attachment:signed_url,
	})
}

func handleSubmitRecomm(resp http.ResponseWriter, req *http.Request){

	user_id,_ :=  public.GetSessionUserId(req)
	user_perm,_ := public.GetSessionUserPermission(req)

	vars := mux.Vars(req)
	hash_str := vars["hash"]

	stagingDb := public.GetNewStagingDatabase()
	defer stagingDb.Session.Close()

	staging_recomm := stagingDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)

	q := staging_recomm.Find(bson.M{
		"hash": hash_str,
	})

	recomm_result := db.RecommEntity{}
	if e := q.One(&recomm_result); e != nil{
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "No such recomm entity",
		})
		return
	}

	// Permission check
	if recomm_result.Recommender != user_id && user_perm != public.USER_PERMISSION_GM {
		public.ResponseStatusAsJson(resp, 403, nil)
		return
	}

	// Migrate from staging db to application db
	staging_recomm.RemoveId(recomm_result.Id)

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	recomm := appDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	new_recomm := recomm_result
	new_recomm.Id = bson.NewObjectId()
	new_recomm.LastModified = time.Now()

	if e := recomm.Insert(new_recomm); e != nil {
		public.LogE.Printf("Error migrating recomm from staging db to application db: %s\n", e.Error())
		public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
			Message: "Submit failed",
		})
	}
}

func handleInspectRecomm(resp http.ResponseWriter, req *http.Request){
	user_perm,_ := public.GetSessionUserPermission(req)

	if !user_perm.ContainsPermission(public.USER_PERMISSION_RECOMM) {
		public.ResponseStatusAsJson(resp, 403, &public.SimpleResult{
			Message: "Not Recommender",
		})
		return
	}

	switch strings.ToLower(req.Method) {

	case "get":{
		// View formal recommendation
		handleViewRecomm(resp, req)
		break;
	}

	case "put":{
		// Submit as formal recommendation
		handleSubmitRecomm(resp, req)
		break;
	}
	case "post":{
		// Submit as formal recommendation
		handleSubmitRecomm(resp, req)
		break;
	}

	default:
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "No such http method",
		})

	}
}

func ConfigRecommHandler(router *mux.Router){
	router.HandleFunc("/doorbell", handleRecommDoorBell)
	router.HandleFunc("/staging", handleViewStagingRecomms)
	router.HandleFunc("/staging/{hash}", handleInspectStagingRecomm)
	router.HandleFunc("/", handleViewFormalRecomm)
	/*
	 Upload handler must either place before /{hash}
	 Or add another path prefix rather than just /upload
	 For the sake of preventing ambiguous
	*/
	router.HandleFunc("/a/upload", handleRecommFileUpload)
	router.HandleFunc("/{hash}", handleInspectRecomm)
}
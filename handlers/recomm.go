package handlers

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/mshockwave/nthuaplus-backend/public"
	"github.com/siddontang/go/bson"
	"github.com/mshockwave/nthuaplus-backend/db"
	"github.com/mshockwave/nthuaplus-backend/storage"
	"time"
	"strings"
)

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

	signed_url := ""
	if client, err := storage.GetNewStorageClient(); err == nil{
		expireTime := time.Now().Add(time.Duration(1) * time.Hour) //an hour
		signed_url,_ = client.GetNewSignedURL(string(recomm_result.Attachment), expireTime)
	}

	public.ResponseOkAsJson(resp, &public.RecommView{
		Hash: recomm_result.Hash,
		ApplyUser: recomm_result.ApplyUser,
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

func ConfigRecommHandler(router *mux.Router){
	router.HandleFunc("/staging", handleViewStagingRecomms)
	router.HandleFunc("/staging/{hash}", handleInspectStagingRecomm)
	router.HandleFunc("/upload", handleRecommFileUpload)
}
package handlers

import (
	"net/http"
	"github.com/mshockwave/nthuaplus-backend/public"
	"github.com/mshockwave/nthuaplus-backend/db"
	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/mux"
	"encoding/json"
	"io/ioutil"
	"time"
	"github.com/mshockwave/nthuaplus-backend/storage"
)

const(
	REVIEWER_DB_RESULT_COLLECTION = "results"
)

var(
	exportAppHashMap map[string]bson.ObjectId
)

func handleGetReviewApplications(resp http.ResponseWriter, req *http.Request){
	userId,_ := public.GetSessionReviewerId(req)

	reviewerDb := public.GetNewReviewerDatabase()
	defer reviewerDb.Session.Close()

	profile := reviewerDb.C(REVIEWER_DB_PROFILE_COLLECTION)

	q := profile.FindId(userId)
	if c, err := q.Count(); c == 0 || err != nil{
		r := public.SimpleResult{
			Message: "Error",
			Description: "User Not Found",
		}
		public.ResponseStatusAsJson(resp, 500, &r)
	}else{
		reviewer := db.Reviewer{}
		q.One(&reviewer)

		appDb := public.GetNewApplicationDatabase()
		defer appDb.Session.Close()

		forms := appDb.C(public.APPLICATION_DB_FORM_COLLECTION)
		results := reviewerDb.C(REVIEWER_DB_RESULT_COLLECTION)

		var exportApps []exportApplication
		for _, t := range reviewer.Topics{
			q := forms.Find(bson.M{
				"topic": t,
			})

			it := q.Iter()
			appData := db.ApplicationForm{}
			for it.Next(&appData) {
				//Check if reviewed
				q_r := results.Find(bson.M{
					"applicationid": appData.Id,
				})
				if n,_ := q_r.Count(); n > 0 {
					//Has reviewed
					continue
				}

				exportApp := exportApplication{}
				(&exportApp).fromDbApplication(&appData, true)
				exportApps = append(exportApps, exportApp)

				exportAppHashMap[exportApp.Hash] = appData.Id
			}
		}

		//Output reviewed topics
		public.ResponseOkAsJson(resp, &exportApps)
	}
}

func handleSubmitReview(resp http.ResponseWriter, req *http.Request){
	vars := mux.Vars(req)
	appHash := vars["appHash"]

	appId, ok := exportAppHashMap[appHash]
	if !ok {
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "Error",
			Description: "Hash not found",
		})
		return
	}
	delete(exportAppHashMap, appHash)

	userId,_ := public.GetSessionReviewerId(req)

	reviewDb := public.GetNewReviewerDatabase()
	defer reviewDb.Session.Close()
	userDb := public.GetNewUserDatabase()
	defer userDb.Session.Close()

	results := reviewDb.C(REVIEWER_DB_RESULT_COLLECTION)
	profiles := reviewDb.C(REVIEWER_DB_PROFILE_COLLECTION)

	//See if exist
	//Re-submit is not allowed
	q := results.Find(bson.M{
		"applicationid": appId,
		"reviewerid": userId,
	})
	if n,_ := q.Count(); n > 0{
		public.ResponseStatusAsJson(resp, 403, &public.SimpleResult{
			Message: "Error",
			Description: "Data exist",
		})
		return
	}

	//Get user profile info
	q = profiles.FindId(userId)
	user := db.User{}
	if err := q.One(&user); err != nil {
		public.ResponseStatusAsJson(resp, 404, nil)
		return
	}

	//Get review json data
	reviewData := public.ReviewResponse{}
	body,_ := ioutil.ReadAll(req.Body)

	if err := json.Unmarshal(body, &reviewData); err != nil {
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
			Description: "Wrong review response",
		})
		return
	}

	reviewResult := db.ReviewResult{
		ApplicationId: appId,
		ReviewerId: userId,
	}
	reviewData.CopyToDbReviewResult(&reviewResult)

	if err := results.Insert(&reviewResult); err != nil {
		public.LogE.Printf("Error inserting new review result: %s\n", err)
	}

	public.ResponseOkAsJson(resp, nil)
}

type reviewerRecommResult struct {
	Content		string
	AttachmentUrl	string ""
}
func handleReviewerRecommView(resp http.ResponseWriter, req *http.Request){
	vars := mux.Vars(req)
	hashStr := vars["recommHash"]

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	recomm := appDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)

	q := recomm.Find(bson.M{
		"hash": hashStr,
	})
	recommInstance := db.Recomm{}
	if err := q.One(&recommInstance); err != nil {
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "Error",
			Description: "Hash not found",
		})
		return
	}

	recommResult := reviewerRecommResult{
		Content: recommInstance.Content,
	}

	if len(recommInstance.Attachment) > 0 {
		//Create temp url
		if client, err := storage.GetNewStorageClient(); err == nil {
			expireTime := time.Now().Add(time.Duration(1) * time.Hour) //an hour
			if obj,e := client.GetNewSignedURL(recommInstance.Attachment, expireTime); e == nil {
				recommResult.AttachmentUrl = obj
			}else{
				public.LogE.Println("Get object error: " + e.Error())
			}
		}
	}

	public.ResponseOkAsJson(resp, &recommResult)
}

func ConfigReviewHandler(router *mux.Router){
	router.HandleFunc("/register", handleReviewRegister)
	router.HandleFunc("/login", handleReviewLogin)
	router.HandleFunc("/logout", public.AuthReviewerVerifyWrapper(handleReviewerLogout))
	router.HandleFunc("/profile", public.AuthReviewerVerifyWrapper(handleReviewerProfile))

	router.HandleFunc("/app",
		public.AuthReviewerVerifyWrapper(handleGetReviewApplications))
	router.HandleFunc("/app/{appHash}",
		public.AuthReviewerVerifyWrapper(public.RequestMethodGuard(handleSubmitReview, "post", "put")))

	router.HandleFunc("/recomm/{recommHash}",
		public.AuthReviewerVerifyWrapper(public.RequestMethodGuard(handleReviewerRecommView, "get")))
}
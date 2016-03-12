package handlers

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/mshockwave/nthuaplus-backend/public"
	"github.com/mshockwave/nthuaplus-backend/db"
	"time"
	"gopkg.in/mgo.v2/bson"
)

const(
	MISC_DB_BULLETIN_COLLECTION = "bulletin"
)

type bulletinNoteResult struct {
	Title		string ""
	Content		string ""
	TimeStamp	time.Time
}
func handleBulletinNotes(resp http.ResponseWriter, req *http.Request){

	miscDb := public.GetNewMiscDatabase()
	defer miscDb.Session.Close()

	bulletin := miscDb.C(MISC_DB_BULLETIN_COLLECTION)
	q := bulletin.Find(bson.M{})
	if _, e := q.Count(); e != nil {
		public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
			Message: "Error",
			Description: "Failed fetching bulletin notes",
		})
		return
	}

	var results []bulletinNoteResult
	it := q.Iter()
	result := db.BulletinNote{}
	for it.Next(&result) {
		results = append(results, bulletinNoteResult{
			Title: result.Title,
			Content: result.Content,
			TimeStamp: result.TimeStamp,
		})
	}

	public.ResponseOkAsJson(resp, &results)
}

type resultAppStatus struct {
	TotalApplicationNum	int

	TopicsNum		[]int
}
func handleApplicationStatus(resp http.ResponseWriter, req *http.Request){
	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	appC := appDb.C(public.APPLICATION_DB_FORM_COLLECTION)
	q := appC.Find(bson.M{})

	result := resultAppStatus{
		TotalApplicationNum: 0,
		TopicsNum: make([]int, len(TOPICS), len(TOPICS)),
	}
	form := db.ApplicationForm{}
	it := q.Iter()
	for it.Next(&form) {
		result.TotalApplicationNum++

		switch form.Topic {
		case 0:
			result.TopicsNum[0] += 1
			break;

		case 1:
			result.TopicsNum[1] += 1
			break;

		case 2:
			result.TopicsNum[2] += 1
			break;

		case 3:
			result.TopicsNum[3] += 1
			break;

		case 4:
			result.TopicsNum[4] += 1
			break;
		}
	}

	public.ResponseOkAsJson(resp, &result)
}

func ConfigMiscHandlers(router *mux.Router){
	router.HandleFunc("/bulletin", public.AuthVerifierWrapper(handleBulletinNotes))

	router.HandleFunc("/status/application", handleApplicationStatus)
}

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

func handleStatus(resp http.ResponseWriter, req *http.Request){

}

func ConfigMiscHandlers(router *mux.Router){
	router.HandleFunc("/bulletin", public.AuthVerifierWrapper(handleBulletinNotes))

	router.HandleFunc("/status", public.AuthVerifierWrapper(handleStatus))
}

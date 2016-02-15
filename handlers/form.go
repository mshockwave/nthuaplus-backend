package handlers

import (
	"github.com/gorilla/mux"
	"net/http"

	"../public"
)

func handleSubmit(resp http.ResponseWriter, req *http.Request){
	//userId,_ := public.GetSessionUserId(req)


}

func ConfigFormHandler(router *mux.Router){
	router.HandleFunc("/submit", public.AuthVerifierWrapper(handleSubmit))
}

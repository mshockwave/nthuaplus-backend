package handlers

import (
	"github.com/gorilla/mux"
	"net/http"

	"../public"
)

func handleReviewRegister(resp http.ResponseWriter, req *http.Request){

}
func handleReviewLogin(resp http.ResponseWriter, req *http.Request) {

}
func handleReviewLogout(resp http.ResponseWriter, req *http.Request){

}

func ConfigReviewHandler(router *mux.Router){
	router.HandleFunc("/register", handleReviewRegister)
	router.HandleFunc("/login", handleReviewLogin)
	router.HandleFunc("/logout", public.AuthVerifierWrapper(handleReviewLogout))
}

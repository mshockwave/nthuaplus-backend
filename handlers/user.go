package handlers

import (
	"net/http"
	"github.com/gorilla/mux"
)

func handleRegister(resp http.ResponseWriter, req *http.Request){

}

func handleLogin(resp http.ResponseWriter, req *http.Request){

}

func GetUserHandler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/register", handleRegister)
	router.HandleFunc("/login", handleLogin)

	return router
}

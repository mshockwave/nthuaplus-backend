package handlers

import (
	"net/http"
	"github.com/gorilla/mux"

	"../public"
	"regexp"
)

func handleRegister(resp http.ResponseWriter, req *http.Request){
	email := req.FormValue("email")
	username := req.FormValue("username")
	formalId := req.FormValue("formalId")
	password := req.FormValue("password")

	//Verify values first
	var errorFields []string
	if len(email) <= 0{ append(errorFields, "Email") }
	if len(username) <= 0{ append(errorFields, "Username") }
	if len(password) <= 0{ append(errorFields, "Password") }
	if len(formalId) != 10{
		append(errorFields, "FormalId")
	}else{
		if match, _ := regexp.MatchString("[A-Z][12][0-9]{8}", formalId); match {
			if !public.FormalIdVerifier(formalId) {
				append(errorFields, "FormalId")
			}
		}else{
			append(errorFields, "FormalId")
		}
	}

	r := public.SimpleResult{
		Message: "This is register",
	}
	public.ResponseAsJson(resp, &r)
}

func handleLogin(resp http.ResponseWriter, req *http.Request){
	r := public.SimpleResult{
		Message: "This is login",
	}
	public.ResponseAsJson(resp, &r)
}

func ConfigUserHandler(router *mux.Router){
	router.HandleFunc("/register", handleRegister)
	router.HandleFunc("/login", handleLogin)
}

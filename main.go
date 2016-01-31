package main

import (
	"github.com/gorilla/mux"
	"net/http"

	"./handlers"
	"./public"
	"fmt"
)

func main() {

	//Setup router
	router := mux.NewRouter()
	router.Handle("/user", handlers.GetUserHandler())

	http.Handle("/", router)

	addrStr := fmt.Sprintf("%s:%d",
		public.Config.GetString("server.address"),
		public.Config.GetInt("server.port"))
	public.LogV.Printf("Listen address: %s\n", addrStr)
	public.LogE.Fatal(http.ListenAndServe(addrStr, nil))
}

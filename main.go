package main

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/mshockwave/nthuaplus-backend/handlers"
	"github.com/mshockwave/nthuaplus-backend/public"
	"fmt"
	"github.com/gorilla/context"
	goHandlers "github.com/gorilla/handlers"
)

func main() {

	//Setup router
	router := mux.NewRouter()
	handlers.ConfigUserHandler(router.PathPrefix("/user").Subrouter())
	handlers.ConfigFormHandler(router.PathPrefix("/form").Subrouter())
	handlers.ConfigMiscHandlers(router.PathPrefix("/misc").Subrouter())
	handlers.ConfigReviewHandler(router.PathPrefix("/review").Subrouter())
	handlers.ConfigRecommHandler(router.PathPrefix("/recomm").Subrouter())
	handlers.ConfigAdminHandler(router.PathPrefix("/gm").Subrouter())

	http.Handle("/", router)

	//Setup CORS Options
	origins := make([]string, 1)
	origins[0] = "*"
	allowOrigins := goHandlers.AllowedOrigins(origins)

	addrStr := fmt.Sprintf("%s:%d",
		public.Config.GetString("server.address"),
		public.Config.GetInt("server.port"))
	public.LogV.Printf("Listen address: %s\n", addrStr)
	public.LogE.Fatal(http.ListenAndServe(
		addrStr,
		context.ClearHandler(goHandlers.CORS(allowOrigins)(http.DefaultServeMux)),
	))
}

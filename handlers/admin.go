package handlers

import(
	"github.com/gorilla/mux"
	"net/http"
	"github.com/mshockwave/nthuaplus-backend/public"
	"github.com/mshockwave/nthuaplus-backend/db"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

const(
	USER_DB_GM_COLLECTION = "gms" //Game Masters, KerKer
)

func handleGMLogin(resp http.ResponseWriter, req *http.Request){
	email := public.EmailFilter(req.FormValue("email"))
	password := req.FormValue("password")

	if len(email) <= 0 || len(password) <= 0 {
		r := public.SimpleResult{
			Message: "Error",
			Description: "Incorrect email or password",
		}
		public.ResponseStatusAsJson(resp, 403, &r)
		return
	}

	//Check login status
	if _, err := public.GetSessionGMId(req); err == nil {
		r := public.SimpleResult{
			Message: "Already Login",
			Description: email,
		}
		public.ResponseOkAsJson(resp, &r)
		return
	}

	userDb := public.GetNewUserDatabase()
	defer userDb.Session.Close()

	profiles := userDb.C(USER_DB_PROFILE_COLLECTION)
	q := profiles.Find( bson.M{"email": email} )
	user := db.User{}
	if q.One(&user) == nil {
		//Check password
		if bcrypt.CompareHashAndPassword([]byte(user.AuthInfo.BcyptHash), []byte(password)) != nil {
			r := public.SimpleResult{
				Message: "Error",
				Description: "Incorrect email or password",
			}
			public.ResponseStatusAsJson(resp, 403, &r)
			return
		}

		//Check whether is GM
		admin := userDb.C(USER_DB_GM_COLLECTION)
		admin_q := admin.Find(bson.M{ "userid": user.Id })
		if n,_ := admin_q.Count(); n <= 0{
			//Not GM
			public.ResponseStatusAsJson(resp, 403, &public.SimpleResult{
				Message: "Error",
				Description: "Not GM, YOU SHALL NOT PASS",
			})
			return
		}

		if err := public.SetGMSessionValue(req, resp, public.GM_ID_SESSION_KEY, user.Id.Hex()); err != nil {
			public.LogE.Printf("Error setting session user id: %s\n", err.Error())
		}
		r := public.SimpleResult{
			Message: "Login Successed",
			Description: email,
		}
		public.ResponseOkAsJson(resp, &r)
	}else{
		r := public.SimpleResult{
			Message: "Error",
			Description: "Incorrect email or password",
		}
		public.ResponseStatusAsJson(resp, 403, &r)
		return
	}
}
func handleGMProfile(resp http.ResponseWriter, req *http.Request){
	user_id,_ := public.GetSessionGMId(req)

	req.Header.Set(public.GM_PERMITTED_HEADER_KEY, user_id.Hex())

	handleUserProfile(resp, req)
}
func handleGMLogout(resp http.ResponseWriter, req *http.Request){
	if err := public.SetGMSessionValue(req, resp, public.GM_ID_SESSION_KEY, nil); err != nil {
		public.LogE.Printf("Logout Failed: %s\n", err.Error())
		public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
			Message: "Error",
			Description: "Logout Failed",
		})
	}else{
		public.ResponseOkAsJson(resp, &public.SimpleResult{
			Message: "Logout Success",
		})
	}
}
func handleQueryAccount(resp http.ResponseWriter, req *http.Request){
	email := req.URL.Query().Get("email")
	if len(email) > 0 {
		userDb := public.GetNewUserDatabase()
		defer userDb.Session.Close()

		profiles := userDb.C(USER_DB_PROFILE_COLLECTION)
		q := profiles.Find(bson.M{ "email": email })
		user := db.User{}
		if e := q.One(&user); e == nil {
			//GM can get more information
			public.ResponseOkAsJson(resp, &user)
		}else{
			public.ResponseStatusAsJson(resp, 404, nil)
		}
	}else{
		public.ResponseStatusAsJson(resp, 404, nil)
	}
}

func gmMockWrapper(resp http.ResponseWriter, req *http.Request, handler http.HandlerFunc){
	email := req.URL.Query().Get("email")
	if len(email) <= 0 {
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
			Description: "Need email",
		})
		return
	}

	userDb := public.GetNewUserDatabase()
	defer userDb.Session.Close()

	profiles := userDb.C(USER_DB_PROFILE_COLLECTION)
	q := profiles.Find(bson.M{"email": email})
	user := db.User{}
	if e := q.One(&user); e != nil {
		public.LogE.Printf("Get controlled user failed: %s\n", e.Error())
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
			Description: "Need email",
		})
		return
	}

	req.Header.Set(public.GM_PERMITTED_HEADER_KEY, user.Id.Hex())
	handler(resp, req)
}

func handleGMFormSubmit(resp http.ResponseWriter, req *http.Request){
	gmMockWrapper(resp, req, handleFormSubmit)
}
func handleGMUploadFile(resp http.ResponseWriter, req *http.Request){
	gmMockWrapper(resp, req, handleFormFileUpload)
}
func handleGMFormView(resp http.ResponseWriter, req *http.Request){
	gmMockWrapper(resp, req, handleFormView)
}

func getUserRecomms(resp http.ResponseWriter, req *http.Request){
	user_id,_ := public.GetSessionUserId(req)

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	forms := appDb.C(public.APPLICATION_DB_FORM_COLLECTION)
	q := forms.Find(bson.M{ "ownerid": user_id }).Sort("-timestamp")
	it := q.Iter()

	form := db.ApplicationForm{}
	var recommList []public.RecommResult
	topicMap := make(map[public.TopicId]bool)
	for it.Next(&form) {
		if _,exist := topicMap[form.Topic]; exist {
			//Skip
			continue
		}

		recomm := appDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)

		for _, h := range form.Recommendations {
			//Transform recommendation from hash to structures
			q := recomm.Find(bson.M{
				"hash": h,
			})

			if n, e := q.Count(); e == nil && n > 0{
				r := db.Recomm{}
				if e := q.One(&r); e == nil {
					r := public.RecommResult{
						Recommender: r.Recommender,
						ApplyUser: r.ApplyUser,
						Done: r.Submitted,
						Hash: h,
					}
					recommList = append(recommList, r)
				}
			}
		}

		topicMap[form.Topic] = true
	}

	public.ResponseOkAsJson(resp, &recommList)
}
func handleGMGetRecomms(resp http.ResponseWriter, req *http.Request){
	gmMockWrapper(resp, req, getUserRecomms)
}

func handleGMRecommsResend(resp http.ResponseWriter, req *http.Request){
	vars := mux.Vars(req)
	hash := vars["recommHash"]

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	recomm := appDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	q := recomm.Find(bson.M{
		"hash": hash,
	})
	recommObj := db.Recomm{}
	if err := q.One(&recommObj); err != nil || len(hash) <= 0{
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "Error",
			Description: "No Such page",
		})
		return
	}

	url := "https://application.nthuaplus.org/recomm.html?hash=" + hash
	if e := public.SendMail(recommObj.Recommender.Email, recommObj.ApplyUser, url); e != nil {
		public.LogE.Printf("Error sending email: %s\n", e.Error())
		public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
			Message: "Error",
		})
	}else{
		public.ResponseOkAsJson(resp, nil)
	}
}

func ConfigAdminHandler(router *mux.Router){
	router.HandleFunc("/login", handleGMLogin)
	router.HandleFunc("/profile", public.AuthGMVerifierWrapper(handleGMProfile))
	router.HandleFunc("/logout", public.AuthGMVerifierWrapper(handleGMLogout))

	router.HandleFunc("/account/profile", public.AuthGMVerifierWrapper(handleQueryAccount))

	router.HandleFunc("/form/submit", public.AuthGMVerifierWrapper(handleGMFormSubmit))
	router.HandleFunc("/form/upload", public.AuthGMVerifierWrapper(handleGMUploadFile))
	router.HandleFunc("/form/view", public.AuthGMVerifierWrapper(handleGMFormView))

	router.HandleFunc("/form/recomm", public.AuthGMVerifierWrapper(handleGMGetRecomms))
	router.HandleFunc("/form/recomm/{recommHash}/resend", public.AuthGMVerifierWrapper(handleGMRecommsResend))
}

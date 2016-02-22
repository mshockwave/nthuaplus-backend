package handlers

import (
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"net/http"

	"../public"
	"../db"
	"../storage"
	"io"
	"regexp"
	"strings"
	"mime/multipart"
	"mime"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const(
	REVIEWER_DB_PROFILE_COLLECTION = "profiles"
)

func handleReviewRegister(resp http.ResponseWriter, req *http.Request){
	email := public.EmailFilter(req.FormValue("email"))
	username := req.FormValue("username")
	formalId := req.FormValue("formalId")
	password := req.FormValue("password")

	//Verify values first
	var errorFields []string
	if len(email) <= 0{ errorFields = append(errorFields, "Email") }
	if len(username) <= 0{ errorFields = append(errorFields, "Username") }
	if len(password) <= 0{ errorFields = append(errorFields, "Password") }
	if len(formalId) != 10{
		errorFields = append(errorFields, "FormalId")
	}else{
		if match, _ := regexp.MatchString("[A-Z][12][0-9]{8}", formalId); match {
			if !public.FormalIdVerifier(formalId) {
				errorFields = append(errorFields, "FormalId")
			}
		}else{
			errorFields = append(errorFields, "FormalId")
		}
	}

	if len(errorFields) > 0 {
		r := public.SimpleResult{
			Message: "Error",
			Description: "Wrong Format: " + strings.Join(errorFields, ","),
		}
		public.ResponseStatusAsJson(resp, 400, &r)
	}else{
		//Get thumbnail if exist
		var thumb multipart.File = nil
		var thumbHeader *multipart.FileHeader = nil
		if f, h, err := req.FormFile("thumbnail"); err == nil && f != nil{
			thumb = f
			thumbHeader = h
		}

		reviewerDb := public.GetNewReviewerDatabase()
		defer reviewerDb.Session.Close()

		profile := reviewerDb.C(REVIEWER_DB_PROFILE_COLLECTION)
		q := profile.Find(bson.M{ "baseprofile.email": email })
		if cnt, err := q.Count(); cnt != 0 || err != nil {
			if err != nil {
				r := public.SimpleResult{
					Message: "Error",
					Description: err.Error(),
				}
				public.ResponseStatusAsJson(resp, 500, &r)
			}else{
				//User exist
				r := public.SimpleResult{
					Message: "Error",
					Description: "User Exists",
				}
				public.ResponseStatusAsJson(resp, 400, &r)
			}
		}else{
			baseUser := db.User{
				Email: email,
				Username: username,
				FormalId: formalId,
			}
			hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			baseUser.AuthInfo = db.UserAuth{
				BcryptCost: bcrypt.DefaultCost,
				BcyptHash: string(hash),
			}

			//Store thumbnail
			if thumb != nil {
				defer thumb.Close()
				if client, err := storage.GetNewStorageClient(); err == nil {

					h := public.NewHashString()
					objName := storage.PathJoin(storage.THUMBNAILS_FOLDER_NAME, h)
					//Determine the extension
					var ext string = ""
					if thumbHeader != nil {
						if segs := strings.Split(thumbHeader.Filename, "."); len(segs) > 1 {
							ext = "." + segs[ len(segs) - 1 ]
							objName = (objName + ext)
						}
					}

					obj := client.GetDefaultBucket().Object(objName)
					if attr, _ := obj.Attrs(client.Ctx); attr != nil {
						if mimeStr := mime.TypeByExtension(ext); len(mimeStr) > 0 {
							attr.ContentType = mimeStr
						}
					}
					objWriter := obj.NewWriter(client.Ctx)
					defer objWriter.Close()

					_, err = io.Copy(objWriter, thumb)
					if err == nil {
						baseUser.Thumbnail = objName
					}
				}
			}

			newUser := db.Reviewer{ {
				BaseProfile: baseUser,
			}}

			if err := profile.Insert(&newUser); err != nil {
				r := public.SimpleResult{
					Message: "Register Failed",
					Description: err.Error(),
				}
				public.ResponseStatusAsJson(resp, 400, &r)
			}else{
				if err := public.SetSessionValue(req, resp, public.USER_ID_SESSION_KEY, newUser.Id.Hex()); err != nil {
					public.LogE.Printf("Error setting session user id: %s\n", err.Error())
				}

				r := public.SimpleResult{
					Message: "Register Successed",
					Description: email,
				}
				public.ResponseOkAsJson(resp, &r)
			}
		}
	}
}
func handleReviewLogin(resp http.ResponseWriter, req *http.Request) {
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
	if _, err := public.GetSessionUserId(req); err == nil {
		r := public.SimpleResult{
			Message: "Already Login",
			Description: email,
		}
		public.ResponseOkAsJson(resp, &r)
		return
	}

	reviewerDb := public.GetNewReviewerDatabase()
	defer reviewerDb.Session.Close()

	profiles := reviewerDb.C(REVIEWER_DB_PROFILE_COLLECTION)
	q := profiles.Find( bson.M{"baseprofile.email": email} )
	reviewer := db.Reviewer{}
	if q.One(&reviewer) == nil {
		//Check password
		if bcrypt.CompareHashAndPassword([]byte(reviewer.BaseProfile.AuthInfo.BcyptHash), []byte(password)) != nil {
			r := public.SimpleResult{
				Message: "Error",
				Description: "Incorrect email or password",
			}
			public.ResponseStatusAsJson(resp, 403, &r)
			return
		}

		if err := public.SetSessionValue(req, resp, public.USER_ID_SESSION_KEY, reviewer.Id.Hex()); err != nil {
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

func handleReviewerProfile(resp http.ResponseWriter, req *http.Request) {
	userId,_ := public.GetSessionUserId(req)

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

		r := public.ReviewerProfile{
			Email: reviewer.BaseProfile.Email,
			Username: reviewer.BaseProfile.Username,
			FormalId: reviewer.BaseProfile.FormalId,

			Topics: reviewer.Topics,
			Permissions: reviewer.Permissions,
		}

		if client,err := storage.GetNewStorageClient(); err == nil && len(reviewer.BaseProfile.Thumbnail) > 0{
			defer client.Close()
			expire := time.Now().Add(time.Duration(12) * time.Hour)
			if r.Thumbnail,err = client.GetNewSignedURL(reviewer.BaseProfile.Thumbnail, expire); err != nil {
				r.Thumbnail = ""
			}
		}

		public.ResponseOkAsJson(resp, &r)
	}
}

func ConfigReviewHandler(router *mux.Router){
	router.HandleFunc("/register", handleReviewRegister)
	router.HandleFunc("/login", handleReviewLogin)
	router.HandleFunc("/logout", public.AuthVerifierWrapper(handleLogout))

	router.HandleFunc("/profile", public.AuthVerifierWrapper(handleReviewerProfile))
}

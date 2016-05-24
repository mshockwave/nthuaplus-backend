package handlers

import (
	"net/http"
	"github.com/gorilla/mux"

	"github.com/mshockwave/nthuaplus-backend/public"
	"github.com/mshockwave/nthuaplus-backend/db"
	"github.com/mshockwave/nthuaplus-backend/storage"
	"regexp"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"golang.org/x/crypto/bcrypt"
	"mime/multipart"
	"io"
	"mime"
	"time"
)

const(
	USER_DB_PROFILE_COLLECTION = "profiles"
)

func handleRegister(resp http.ResponseWriter, req *http.Request){
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

		userDb := public.GetNewUserDatabase()
		defer userDb.Session.Close()

		profile := userDb.C(USER_DB_PROFILE_COLLECTION)
		q := profile.Find(bson.M{ "email": email })
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
			newUser := db.User{
				Email: email,
				Username: username,
				FormalId: formalId,
			}
			hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			newUser.AuthInfo = db.UserAuth{
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
						newUser.Thumbnail = objName
					}
				}
			}

			if err := profile.Insert(&newUser); err != nil {
				r := public.SimpleResult{
					Message: "Register Failed",
					Description: err.Error(),
				}
				public.ResponseStatusAsJson(resp, 400, &r)
			}else{
				if err := public.SetUserSessionValue(req, resp, public.USER_ID_SESSION_KEY, newUser.Id.Hex()); err != nil {
					public.LogE.Printf("Error setting session user id: %s\n", err.Error())
				}

				if err := public.SetUserSessionValue(req, resp, public.USER_PERMISSION_SESSION_KEY, newUser.Permission); err != nil {
					public.LogE.Printf("Error setting session user permission: %s\n", err.Error())
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

func handleLogin(resp http.ResponseWriter, req *http.Request){
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

		if err := public.SetUserSessionValue(req, resp, public.USER_ID_SESSION_KEY, user.Id.Hex()); err != nil {
			public.LogE.Printf("Error setting session user id: %s\n", err.Error())
		}

		if err := public.SetUserSessionValue(req, resp, public.USER_PERMISSION_SESSION_KEY, user.Permission); err != nil {
			public.LogE.Printf("Error setting session user permission: %s\n", err.Error())
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

func handleLogout(resp http.ResponseWriter, req *http.Request){

	if e := public.SetUserSessionValue(req, resp, public.USER_PERMISSION_SESSION_KEY, nil); e != nil {
		public.LogE.Printf("Error cleaning session user permission: %s\n", e.Error())
	}

	if err := public.SetUserSessionValue(req, resp, public.USER_ID_SESSION_KEY, nil); err != nil {
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

func handleUserProfile(resp http.ResponseWriter, req *http.Request){
	userId,_ := public.GetSessionUserId(req)

	userDb := public.GetNewUserDatabase()
	defer userDb.Session.Close()

	profile := userDb.C(USER_DB_PROFILE_COLLECTION)

	q := profile.FindId(userId)
	if c, err := q.Count(); c == 0 || err != nil{
		r := public.SimpleResult{
			Message: "Error",
			Description: "User Not Found",
		}
		public.ResponseStatusAsJson(resp, 500, &r)
	}else{
		user := db.User{}
		q.One(&user)

		r := public.UserProfile{
			Email: user.Email,
			Username: user.Username,
			FormalId: user.FormalId,
		}

		if client,err := storage.GetNewStorageClient(); err == nil && len(user.Thumbnail) > 0{
			defer client.Close()
			expire := time.Now().Add(time.Duration(12) * time.Hour)
			if r.Thumbnail,err = client.GetNewSignedURL(user.Thumbnail, expire); err != nil {
				r.Thumbnail = ""
			}
		}

		public.ResponseOkAsJson(resp, &r)
	}
}

func ConfigUserHandler(router *mux.Router){
	router.HandleFunc("/register", handleRegister)
	router.HandleFunc("/login", handleLogin)
	router.HandleFunc("/logout", public.AuthUserVerifierWrapper(handleLogout))

	router.HandleFunc("/profile", public.AuthUserVerifierWrapper(handleUserProfile))
}

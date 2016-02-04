package handlers

import (
	"net/http"
	"github.com/gorilla/mux"

	"../public"
	"../db"
	"../storage"
	"regexp"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"golang.org/x/crypto/bcrypt"
	"mime/multipart"
	"io"
	"mime"
)

const(
	USER_DB_PROFILE_COLLECTION = "profiles"
)

func handleRegister(resp http.ResponseWriter, req *http.Request){
	email := req.FormValue("email")
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
		public.ResponseAsJson(resp, &r)
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
				public.ResponseAsJson(resp, &r)
			}else{
				//User exist
				r := public.SimpleResult{
					Message: "User Exist",
					Description: email,
				}
				public.ResponseAsJson(resp, &r)
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
				public.ResponseAsJson(resp, &r)
			}else{
				r := public.SimpleResult{
					Message: "Register Successed",
					Description: email,
				}
				public.ResponseAsJson(resp, &r)
			}
		}
	}
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

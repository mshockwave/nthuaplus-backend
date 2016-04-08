package handlers

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/mshockwave/nthuaplus-backend/public"
	"github.com/mshockwave/nthuaplus-backend/db"
	"github.com/mshockwave/nthuaplus-backend/storage"
	"github.com/wendal/errors"
	"fmt"
	"time"
	"encoding/json"
	"strings"
	"io"
	"mime/multipart"
	"mime"
	"gopkg.in/mgo.v2/bson"
)

var(
	TOPICS = []string{"topic1", "topic2", "topic3", "topic4", "topic5"}
)

func handleFormSubmit(resp http.ResponseWriter, req *http.Request){
	ownerId,_ := public.GetSessionUserId(req)

	form := db.ApplicationForm{
		OwnerId: ownerId,
		Timestamp: time.Now(),

		Name: req.FormValue("name"),
		School: req.FormValue("school"),
		Department: req.FormValue("department"),
		Email: req.FormValue("email"),
		Phone: req.FormValue("phoneNumber"),
		Address: req.FormValue("address"),
		FormalId: req.FormValue("formalId"), //TODO: Verify

		Teacher: req.FormValue("teacher"),
		ResearchArea: req.FormValue("researchArea"),
		RelatedSkills: req.FormValue("relatedSkills"),

		ResearchPlan: req.FormValue("researchPlan"),
		Transcript: req.FormValue("transcript"),
		Others: req.FormValue("others"),
	}

	if topic, err := parseTopic(req); err != nil {
		public.LogE.Println(err.Error())
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
			Description: "Wrong form format",
		})
		return
	}else{
		form.Topic = db.TopicId(topic)
	}

	if grade, err := parseSchoolGrade(req); err != nil{
		public.LogE.Println(err.Error())
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
			Description: "Wrong form format",
		})
		return
	}else{
		form.SchoolGrade = grade
	}

	if birthday, err := parseBirthday(req); err != nil{
		public.LogE.Println(err.Error())
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
			Description: "Wrong form format",
		})
		return
	}else{
		form.Birthday = birthday
	}

	if classes, err := parseStudiedClasses(req); err != nil{
		public.LogE.Println(err.Error())
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
			Description: "Wrong form format",
		})
		return
	}else{
		form.ClassHistories = classes
	}

	if languages, err := parseLanguageAbility(req); err != nil{
		public.LogE.Println(err.Error())
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
			Description: "Wrong form format",
		})
		return
	}else{
		form.LangAbilities = languages
	}

	if average, ranking, err := parseAcademicGrades(req); err != nil {
		public.LogE.Println(err.Error())
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
			Description: "Wrong form format",
		})
		return
	}else{
		form.AcademicGrade = db.AcademicGrade{
			Average: average,
			Rank: ranking,
		}
	}

	if letters, err := parseRecommendationLetters(req); err != nil{
		public.LogE.Println(err.Error())
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
			Description: "Wrong form format",
		})
		return
	}else{
		form.Recommendations = handleRecommendationLetters(letters, form.Name, form.Email)
	}


	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	forms := appDb.C(public.APPLICATION_DB_FORM_COLLECTION)
	if err := forms.Insert(&form); err != nil {
		public.LogE.Printf("Insert new form error: " + err.Error())
		public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
			Message: "Error",
			Description: "Add Form Error",
		})
	}else{
		public.ResponseOkAsJson(resp, &public.SimpleResult{
			Message: "Success",
		})
	}
}
func parseTopic(req *http.Request) (uint, error){
	for i,v := range TOPICS {
		if len(req.FormValue(v)) == 0 {
			continue
		}
		return uint(i),nil
	}
	return 1, errors.New("Not Found")
}
func parseSchoolGrade(req *http.Request) (string,error) {
	gradeType := req.FormValue("gradeType")
	schoolGrade := req.FormValue("schoolGrade")

	var numGrade int
	if n,_ := fmt.Sscanf(schoolGrade, "%d", &numGrade); n < 1 || numGrade < 0{
		return "", errors.New("Invalid Grade")
	}

	return fmt.Sprintf("%s@%d", gradeType, numGrade), nil
}
func parseBirthday(req *http.Request) (time.Time, error){
	return time.Parse("2006-01-02", req.FormValue("birthday"))
}
func parseStudiedClasses(req *http.Request) ([]db.StudiedClass, error) {
	var classes []db.StudiedClass

	rawJson := req.FormValue("classHistory")
	if len(rawJson) == 0{
		return classes,errors.New("No argument")
	}

	decoder := json.NewDecoder(strings.NewReader(rawJson))
	if _,e := decoder.Token(); e != nil{ //The first array bracket
		return classes,errors.New("Wrong json format")
	}

	element := db.StudiedClass{}
	for decoder.More() {
		if e := decoder.Decode(&element); e != nil {
			continue
		}
		classes = append(classes, element)
	}

	decoder.Token() //The last array bracket
	return classes,nil
}
func parseRecommendationLetters(req *http.Request) ([]db.BasicUser, error){
	var letters []db.BasicUser

	rawJson := req.FormValue("recommendationLetters")
	if len(rawJson) == 0 {
		return letters, errors.New("No argument")
	}

	decoder := json.NewDecoder(strings.NewReader(rawJson))
	if _,e := decoder.Token(); e != nil {//The first array bracket
		return letters,errors.New("Wrong json format")
	}

	element := db.BasicUser{}
	for decoder.More() {
		if e := decoder.Decode(&element); e != nil {
			continue
		}
		letters = append(letters, element)
	}

	decoder.Token() //The last array bracket
	return letters,nil
}
func handleRecommendationLetters(letters []db.BasicUser, name, email string) []string{
	var hashList []string

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	recomm := appDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	for _, l := range letters{
		r := db.Recomm{

			Hash: public.NewSecureHashString(),

			Submitted: false,

			ApplyUser: db.BasicUser{
				Name: name,
				Email: email,
			},
			Recommender: db.BasicUser{
				Name: l.Name,
				Email: l.Email,
			},
		}
		if e := recomm.Insert(&r); e != nil {
			public.LogE.Printf("Failed inserting recommendation entity for applyer %s", name)
		}else{
			hashList = append(hashList, r.Hash)

			url := "https://application.nthuaplus.org/recomm.html?hash=" + r.Hash

			applier := r.ApplyUser
			applier.Name = public.ConvertName(applier.Name)
			if e := public.SendMail(l.Email, applier, url); e != nil {
				public.LogE.Println("Error sending email to " + l.Email + ": " + e.Error())
			}
		}
	}

	return hashList
}

type RawLang struct {
	LangName string
	Abilities	[]RawLangAbility
}
type RawLangAbility struct {
	Text	string
	Value	float64
}
func parseLanguageAbility(req *http.Request) ([]db.LanguageAbility,error) {
	var languages []db.LanguageAbility

	rawJson := req.FormValue("langAbilities")
	if len(rawJson) == 0{
		return languages,errors.New("No argument")
	}

	decoder := json.NewDecoder(strings.NewReader(rawJson))
	if _,e := decoder.Token(); e != nil{ //The first array bracket
		return languages,errors.New("Wrong json format")
	}

	element := RawLang{}
	for decoder.More() {
		if e := decoder.Decode(&element); e != nil {
			continue
		}

		if len(element.Abilities) < 4 {
			continue
		}
		lang := db.LanguageAbility{
			Name: element.LangName,
			Listening: uint(element.Abilities[0].Value),
			Speaking: uint(element.Abilities[1].Value),
			Reading: uint(element.Abilities[2].Value),
			Writing: uint(element.Abilities[3].Value),
		}
		languages = append(languages, lang)
	}

	decoder.Token() //The last array bracket
	return languages,nil
}
func parseAcademicGrades(req *http.Request) (db.GradeType, db.RankType, error){
	average := req.FormValue("average")
	ranking := req.FormValue("ranking")

	var averageNum float64
	var rankingNum int32

	if n,_ := fmt.Sscanf(average, "%f", &averageNum); n < 1{
		averageNum = float64(-1)
	}
	if n,_ := fmt.Sscanf(ranking, "%d", &rankingNum); n < 1{
		rankingNum = int32(-1)
	}

	return db.GradeType(averageNum), db.RankType(rankingNum), nil
}

func handleFormFileUpload(resp http.ResponseWriter, req *http.Request) {
	if f,h,e := req.FormFile("file"); e == nil && f != nil && h != nil{
		if objName, err := saveFile(h, f); err == nil {
			public.ResponseOkAsJson(resp, &public.SimpleResult{
				Message: "Success",
				Description: objName,
			})
		}else{
			public.LogE.Printf("Error storing file: %s\n", err.Error())
			public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
				Message: "Error",
			})
		}
	}else{
		public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
			Message: "Error",
		})
	}
}

//File saving routines
func saveFile(header *multipart.FileHeader, r io.Reader) (string, error) {
	if client, err := storage.GetNewStorageClient(); err == nil {
		h := public.NewHashString()
		objName := storage.PathJoin(storage.APPLICATIONS_FOLDER_NAME, h)
		//Determine the extension
		var ext string = ""
		if header != nil {
			if segs := strings.Split(header.Filename, "."); len(segs) > 1 {
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

		_, err = io.Copy(objWriter, r)
		if err == nil {
			return objName,nil
		}else{
			return "",err
		}
	}else{
		return "",err
	}
}

type exportApplication struct {
	Timestamp       time.Time
	Hash		string "" //Optional

	//Basic Data
	Name            string
	School          string
	Department      string
	SchoolGrade     string
	Birthday        time.Time
	FormalId        string
	Phone           string
	Email           string
	Address         string

	//Academic Data
	Topic           db.TopicId
	Teacher         string
	ResearchArea    string
	ClassHistories  []db.StudiedClass
	RelatedSkills   string
	AcademicGrade   db.AcademicGrade
	LangAbilities   []db.LanguageAbility

	ResearchPlan    string //File
	Recommendations []public.RecommResult
	Transcript      string //File
	Others          string //File
}
func (this *exportApplication) fromDbApplication(form *db.ApplicationForm, isReviewer bool){
	this.Timestamp = form.Timestamp

	this.Name = form.Name
	this.School = form.School
	this.Department = form.Department
	this.SchoolGrade = form.SchoolGrade
	this.Birthday = form.Birthday
	this.FormalId = form.FormalId
	this.Phone = form.Phone
	this.Email = form.Email
	this.Address = form.Address

	this.Topic = form.Topic
	this.Teacher = form.Teacher
	this.ResearchArea = form.ResearchArea
	this.ClassHistories = form.ClassHistories
	this.RelatedSkills = form.RelatedSkills
	this.AcademicGrade = form.AcademicGrade
	this.LangAbilities = form.LangAbilities

	//Hash
	this.Hash = public.NewSecureHashString()

	//Extras
	//Transform file id to url
	if client,err := storage.GetNewStorageClient(); err == nil {

		expireTime := time.Now().Add(time.Duration(1) * time.Hour) //an hour
		if obj,e := client.GetNewSignedURL(form.ResearchPlan, expireTime); e == nil {
			this.ResearchPlan = obj
		}else{
			public.LogE.Println("Get object error: " + e.Error())
		}
		if obj,e := client.GetNewSignedURL(form.Transcript, expireTime); e == nil {
			this.Transcript = obj
		}else{
			public.LogE.Println("Get object error: " + e.Error())
		}
		if len(form.Others) > 0 {
			if obj,e := client.GetNewSignedURL(form.Others, expireTime); e == nil {
				this.Others = obj
			}else{
				public.LogE.Println("Get object error: " + e.Error())
			}
		}
	}else{
		public.LogE.Printf("Error getting storage client")
	}

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()
	recomm := appDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	var recommList []public.RecommResult

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
				}
				if isReviewer {
					r.Hash = h
				}else{
					r.Hash = ""
				}
				recommList = append(recommList, r)
			}
		}
	}
	this.Recommendations = recommList
}
func handleFormView(resp http.ResponseWriter, req *http.Request) {
	userId,_ := public.GetSessionUserId(req)

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	forms := appDb.C(public.APPLICATION_DB_FORM_COLLECTION)
	q := forms.Find(bson.M{
		"ownerid": userId,
	})
	if _, e := q.Count(); e != nil {
		public.LogE.Println("Query user form error: " + e.Error());
		public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
			Message: "Error",
		})
	}else{

		var formResults []exportApplication
		form := db.ApplicationForm{}
		it := q.Iter()

		for it.Next(&form) {
			exportForm := exportApplication{}
			(&exportForm).fromDbApplication(&form, false)
			formResults = append(formResults, exportForm)
		}

		public.ResponseOkAsJson(resp, formResults)

		/*
		if client, err := storage.GetNewStorageClient(); err == nil {
			defer client.Close()
			var formResults []db.ApplicationForm
			form := db.ApplicationForm{}

			it := q.Iter()
			expireTime := time.Now().Add(time.Duration(1) * time.Hour) //an hour
			for it.Next(&form) {
				form.Id = bson.ObjectId("")
				form.OwnerId = bson.ObjectId("")

				//Handle the file objects
				if obj,e := client.GetNewSignedURL(form.ResearchPlan, expireTime); e == nil {
					form.ResearchPlan = obj
				}else{
					public.LogE.Println("Get object error: " + e.Error())
				}
				if obj,e := client.GetNewSignedURL(form.Transcript, expireTime); e == nil {
					form.Transcript = obj
				}else{
					public.LogE.Println("Get object error: " + e.Error())
				}
				if len(form.Others) > 0 {
					if obj,e := client.GetNewSignedURL(form.Others, expireTime); e == nil {
						form.Others = obj
					}else{
						public.LogE.Println("Get object error: " + e.Error())
					}
				}

				formResults = append(formResults, form)
			}

			public.ResponseOkAsJson(resp, formResults)
		}else{
			public.LogE.Println("Error getting storage client: " + err.Error())
			public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
				Message: "Error",
			})
		}
		*/
	}
}

/**
	POST/PUT:	submit
	GET(or else):	Get info
**/
func handleRecommendation(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	hash := vars["hash"]

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	recomm := appDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	q := recomm.Find(bson.M{
		"hash": hash,
	})
	if n,err := q.Count(); err != nil || n <= 0  || len(hash) <= 0{
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "Error",
			Description: "No Such page",
		})
		return
	}

	result := db.Recomm{}
	if err := q.One(&result); err != nil {
		public.LogE.Printf("Error fetching recommendation data for %s\n", hash)
		public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
			Message: "Error",
		})
		return
	}

	if req.Method == "POST" || req.Method == "PUT" {
		//Submit
		if result.Submitted { //Already submitted
			public.ResponseStatusAsJson(resp, 400, &public.SimpleResult{
				Message: "Error",
			})
			return
		}

		textContent := req.FormValue("textContent")
		fileObj := req.FormValue("fileObj")

		result.Content = textContent
		result.Attachment = fileObj
		result.Submitted = true
		err := recomm.UpdateId(result.Id, &result)
		if err != nil {
			public.LogE.Println("Update recommendation fields error: " + err.Error())
			public.ResponseStatusAsJson(resp, 500, &public.SimpleResult{
				Message: "Error",
			})
		}else{
			public.ResponseOkAsJson(resp, &public.SimpleResult{
				Message: "Success",
			})
		}
	}else{
		//Get info
		displayResult := public.RecommResult{
			Recommender: result.Recommender,
			ApplyUser: result.ApplyUser,
			Done: result.Submitted,
		}
		public.ResponseOkAsJson(resp, &displayResult)
	}
}

func handleRecommendationUpload(resp http.ResponseWriter, req *http.Request){
	vars := mux.Vars(req)
	hash := vars["hash"]

	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	recomm := appDb.C(public.APPLICATION_DB_RECOMM_COLLECTION)
	q := recomm.Find(bson.M{
		"hash": hash,
	})
	if n,err := q.Count(); err != nil || n <= 0  || len(hash) <= 0{
		public.ResponseStatusAsJson(resp, 404, &public.SimpleResult{
			Message: "Error",
			Description: "No Such page",
		})
		return
	}

	handleFormFileUpload(resp, req)
}

func ConfigFormHandler(router *mux.Router){
	router.HandleFunc("/submit", public.AuthUserVerifierWrapper(handleFormSubmit))
	router.HandleFunc("/upload", public.AuthUserVerifierWrapper(handleFormFileUpload))
	router.HandleFunc("/view", public.AuthUserVerifierWrapper(handleFormView))

	router.HandleFunc("/recomm/{hash}", handleRecommendation)
	router.HandleFunc("/recomm/{hash}/upload", public.RequestMethodGuard(handleRecommendationUpload, "post", "put"))
}

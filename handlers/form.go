package handlers

import (
	"github.com/gorilla/mux"
	"net/http"

	"../public"
	"../db"
	"../storage"
	"github.com/wendal/errors"
	"fmt"
	"time"
	"encoding/json"
	"strings"
	"io"
	"mime/multipart"
	"mime"
)


const(
	APPLICATION_DB_FORM_COLLECTION = "forms"
)
var(
	TOPICS = []string{"topic1", "topic2", "topic3"}
)

func handleSubmit(resp http.ResponseWriter, req *http.Request){
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
		Recommendations: req.FormValue("recommendationLetters"),
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
		form.Topic = topic
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


	appDb := public.GetNewApplicationDatabase()
	defer appDb.Session.Close()

	forms := appDb.C(APPLICATION_DB_FORM_COLLECTION)
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
	if n,_ := fmt.Sscanf(average + " " + ranking, "%f %d", &averageNum, &rankingNum); n < 2{
		return db.GradeType(0.0), db.RankType(0), errors.New("Error academic grades format")
	}

	return db.GradeType(averageNum), db.RankType(rankingNum), nil
}

func handleUploadFile(resp http.ResponseWriter, req *http.Request) {
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

func ConfigFormHandler(router *mux.Router){
	router.HandleFunc("/submit", public.AuthVerifierWrapper(handleSubmit))
	router.HandleFunc("/upload", public.AuthVerifierWrapper(handleUploadFile))
}

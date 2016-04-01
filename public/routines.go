package public

import (
	"net/http"
	"encoding/json"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"github.com/wendal/errors"
	"github.com/dchest/uniuri"
	gomail "gopkg.in/gomail.v2"
	"github.com/mshockwave/nthuaplus-backend/db"
	"bytes"
)

func ResponseOkAsJson(resp http.ResponseWriter, value interface{}) (int, error){
	return ResponseStatusAsJson(resp, 200, value)
}
func ResponseStatusAsJson(resp http.ResponseWriter, status int, value interface{}) (int, error){
	if value == nil {
		//Just response the status code
		resp.WriteHeader(status)
		return status, nil
	}

	if j_bytes, err := json.Marshal(value); err != nil {
		resp.WriteHeader(500)
		return 500, err
	}else{
		//Restore '&'
		str := string(j_bytes)
		str = strings.Replace(str, `\u0026`, "&", -1)

		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		resp.WriteHeader(status)
		_, err = resp.Write([]byte(str))
		return status, err
	}
}

func FormalIdVerifier(str string) bool {
	num1 := -1
	num2 := 0
	for i, ch := range str {
		if i == 0{
			switch ch {
			case 'A':
				num1 = 10
				break
			case 'B':
				num1 = 11
				break;
			case 'C':
				num1 = 12
				break;
			case 'D':
				num1 = 13
				break;
			case 'E':
				num1 = 14
				break;
			case 'F':
				num1 = 15
				break;
			case 'G':
				num1 = 16
				break;
			case 'H':
				num1 = 16
				break;
			case 'I':
				num1 = 34
				break;
			case 'J':
				num1 = 18
				break;
			case 'K':
				num1 = 19
				break;
			case 'L':
				num1 = 20
				break;
			case 'M':
				num1 = 21
				break;
			case 'N':
				num1 = 22
				break;
			case 'O':
				num1 = 35
				break;
			case 'P':
				num1 = 23
				break;
			case 'Q':
				num1 = 24
				break;
			case 'R':
				num1 = 25
				break;
			case 'S':
				num1 = 26
				break;
			case 'T':
				num1 = 27
				break;
			case 'U':
				num1 = 28
				break;
			case 'V':
				num1 = 29
				break;
			case 'W':
				num1 = 32
				break;
			case 'X':
				num1 = 30
				break;
			case 'Y':
				num1 = 31
				break;
			case 'Z':
				num1 = 33
				break;
			}
			if num1 < 0 {return false}
			num1 = (num1 % 10) * 9 + (num1 / 10)
		}else{
			var v int = int(ch) - int('0')
			if i < 8 {
				num2 += (9 - i) * v
			}else{
				num2 += v
			}
		}
	}

	return ((num1 + num2) % 10) == 0
}

func EmailFilter(orig string) string { return strings.Replace(orig, "%40", "@", -1) }

func StringJoin(sep string, elements ...string) string{ return strings.Join(elements, sep) }

func NewHashString() string { return bson.NewObjectId().Hex() }
func NewSecureHashString() string { return uniuri.New() }

func ConvertName(rawString string) string { return strings.Replace(rawString, "#", "/", -1) }

func GetSessionValue(req *http.Request, key interface{}) (interface{}, error) {
	s, err := SessionStorage.Get(req, USER_AUTH_SESSION)
	if err != nil { return nil, err }

	return s.Values[key], nil
}
func GetReviewerSessionValue(req *http.Request, key interface{}) (interface{}, error) {
	s, err := SessionStorage.Get(req, REVIEWER_AUTH_SESSION)
	if err != nil { return nil, err }

	return s.Values[key], nil
}
func SetSessionValue(req *http.Request, resp http.ResponseWriter, key, value interface{}) error {
	//Ignore the error since sometimes the browser side coolie storage is broken
	//But we still can assign new cookies
	s, _ := SessionStorage.Get(req, USER_AUTH_SESSION)
	if s == nil { return errors.New("Session " + USER_AUTH_SESSION + " not available") }

	s.Values[key] = value
	return s.Save(req, resp)
}
func SetReviewerSessionValue(req *http.Request, resp http.ResponseWriter, key, value interface{}) error {
	//Ignore the error since sometimes the browser side coolie storage is broken
	//But we still can assign new cookies
	s, _ := SessionStorage.Get(req, REVIEWER_AUTH_SESSION)
	if s == nil { return errors.New("Session " + REVIEWER_AUTH_SESSION + " not available") }

	s.Values[key] = value
	return s.Save(req, resp)
}

func GetSessionUserId(req *http.Request) (bson.ObjectId, error){
	if v, err := GetSessionValue(req, USER_ID_SESSION_KEY); err != nil || v == nil{
		return bson.ObjectId(""), errors.New("Invalid session id format")
	}else{
		if str, found := v.(string); found {
			if bson.IsObjectIdHex(str) {
				return bson.ObjectIdHex(str), nil
			}else{
				return bson.ObjectId(""), errors.New("Invalid session id format")
			}
		}else{
			return bson.ObjectId(""), errors.New("Invalid session id format")
		}
	}
}
func GetSessionReviewerId(req *http.Request) (bson.ObjectId, error){
	if v, err := GetReviewerSessionValue(req, REVIEWER_ID_SESSION_KEY); err != nil || v == nil{
		return bson.ObjectId(""), errors.New("Invalid session id format")
	}else{
		if str, found := v.(string); found {
			if bson.IsObjectIdHex(str) {
				return bson.ObjectIdHex(str), nil
			}else{
				return bson.ObjectId(""), errors.New("Invalid session id format")
			}
		}else{
			return bson.ObjectId(""), errors.New("Invalid session id format")
		}
	}
}

func AuthVerifierWrapper(handler http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request){
		if _, err := GetSessionUserId(req); err != nil {
			r := SimpleResult{
				Message: "Error",
				Description: "Please Login First",
			}
			ResponseStatusAsJson(resp, 403, &r)
			return
		}

		handler(resp, req)
	}
}
func AuthReviewerVerifyWrapper(handler http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request){
		if _, err := GetSessionReviewerId(req); err != nil {
			r := SimpleResult{
				Message: "Error",
				Description: "Please Login First",
			}
			ResponseStatusAsJson(resp, 403, &r)
			return
		}

		handler(resp, req)
	}
}
func RequestMethodGuard(handler http.HandlerFunc, methods ...string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request){
		match := false
		for _,method := range methods {
			method = strings.ToLower(method)
			if strings.ToLower(req.Method) == method {
				match = true
				break;
			}
		}

		if !match {
			ResponseStatusAsJson(resp, 404, &SimpleResult{
				Message: "Error",
				Description: req.Method + " is not supported",
			})
			return
		}

		handler(resp, req)
	}
}

func SendMail(to string, applier db.BasicUser, url string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", "noreply@nthuaplus.org")
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", RecommLetterSubject)

	//Write letter template
	data := struct{
		ApplyUser db.BasicUser
		RecommUrl string
	}{
		ApplyUser: applier,
		RecommUrl: url,
	}
	buffer := new(bytes.Buffer)
	if e := RecommLetterTmpl.Execute(buffer, &data); e != nil {
		return e
	}
	msg.SetBody("text/plain", buffer.String())

	dialer := gomail.Dialer{
		Host: "localhost",
		Port: 25,
	}
	return dialer.DialAndSend(msg)
}
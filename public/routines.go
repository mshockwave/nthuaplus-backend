package public

import (
	"net/http"
	"encoding/json"
	"strings"
	"gopkg.in/mgo.v2/bson"
)

func ResponseAsJson(resp http.ResponseWriter, value interface{}) (int, error){
	if j_bytes, err := json.Marshal(value); err != nil {
		resp.WriteHeader(500)
		return 500, err
	}else{
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, err = resp.Write(j_bytes)
		return 200, err
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

func StringJoin(sep string, elements ...string) string{ return strings.Join(elements, sep) }

func NewHashString() string { return bson.NewObjectId().Hex() }
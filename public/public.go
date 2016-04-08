package public

import (
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"io"
	"os"
	"log"
	"io/ioutil"
	"github.com/gorilla/sessions"
	"github.com/wendal/errors"
	"text/template"
)

const(
	MAIN_STORAGE_BUCKET = "nthu-a-plus-storage"

	USER_AUTH_SESSION = "user-auth"
	USER_ID_SESSION_KEY = "user_id"
	REVIEWER_AUTH_SESSION = "reviewer-auth"
	REVIEWER_ID_SESSION_KEY = "reviewer_id"
	GM_AUTH_SESSION = "gm-auth"
	GM_ID_SESSION_KEY = "gm_id"

	APPLICATION_DB_FORM_COLLECTION = "forms"
	APPLICATION_DB_RECOMM_COLLECTION = "recomms"
)

var(

	CONFIG_FILE_NAME string = "config"
	Config	*viper.Viper

	DB_ADDRESS string = "127.0.0.1"

	mainDbSession *mgo.Session
	userDbSession *mgo.Session
	applicationDbSession *mgo.Session
	miscDbSession *mgo.Session
	reviewerDbSession *mgo.Session

	//Loggers
	LogV	*log.Logger
	LogD	*log.Logger
	LogE	*log.Logger
	LogW	*log.Logger

	//Session
	SessionStorage *sessions.CookieStore

	//Cloud Storage Signed URL
	StoragePrivateKey []byte
	StorageServiceAccountEmail string

	RecommLetterSubject	string
	RecommLetterTmpl	*template.Template
)

func init(){
	if e := initConfig(); e != nil {
		log.Fatalln("Error reading config file: " + e.Error())
		panic(e)
	}

	initLoggers()

	if e := initDatabases(); e != nil {
		panic(e)
	}

	initSession()

	initStorage()

	initRecommLetter()
}

func setDefaultValues(){
	Config.SetDefault("log.filePath", "")
	Config.SetDefault("log.enableStdOut", false)
	Config.SetDefault("log.enableStdErr", false)

	Config.SetDefault("server.address", "")
	Config.SetDefault("server.port", 8888)
}
func initConfig() error{
	Config = viper.New()
	Config.SetConfigName(CONFIG_FILE_NAME)
	Config.AddConfigPath(".")
	Config.AddConfigPath("..")

	setDefaultValues()

	return Config.ReadInConfig()
}

func initLoggers() {
	var writer io.Writer = ioutil.Discard
	var errWriter io.Writer = ioutil.Discard

	if Config.GetBool("log.enableStdOut") {
		writer = io.MultiWriter(writer, os.Stdout)
	}
	if Config.GetBool("log.enableStdErr") {
		errWriter = io.MultiWriter(errWriter, os.Stderr)
	}

	logFilePath := Config.GetString("log.filePath")
	if len(logFilePath) > 0 {
		if file, err := os.Open(logFilePath); err == nil {
			writer = io.MultiWriter(writer, file)
			errWriter = io.MultiWriter(errWriter, file)
		}
	}

	LogV = log.New(writer, "[VERBOSE]:", log.Ldate | log.Ltime | log.Lshortfile)
	LogD = log.New(writer, "[DEBUG]:", log.Ldate | log.Ltime | log.Lshortfile)
	LogE = log.New(errWriter, "[ERROR]:", log.Ldate | log.Ltime | log.Lshortfile)
	LogW = log.New(errWriter, "[WARNING]:", log.Ldate | log.Ltime | log.Lshortfile)

	//fmt.Printf("Log enable stdout: %v\n", Config.GetBool("log.enableStdOut"))
	//fmt.Printf("Log enable stderr: %v\n", Config.GetBool("log.enableStdErr"))
}

func initDatabases() error {
	var err error

	if Config.IsSet("db.address") {
		DB_ADDRESS = Config.GetString("db.address")
	}

	if mainDbSession, err = mgo.Dial(DB_ADDRESS); err != nil {
		LogE.Println("Error connecting database: " + err.Error())
		return err
	}

	if !Config.IsSet("db.username") || !Config.IsSet("db.password") {
		LogE.Println("No database credential")
		return nil
	}
	username := Config.GetString("db.username")
	password := Config.GetString("db.password")

	//Init user session
	userDbSession = mainDbSession.Copy()
	err = userDbSession.Login(&mgo.Credential{
		Username: username,
		Password: password,
		Source: "users",
	})
	if err != nil {
		LogE.Println("User database login failed: " + err.Error())
		return err
	}

	//Init application session
	applicationDbSession = mainDbSession.Copy()
	err = applicationDbSession.Login(&mgo.Credential{
		Username: username,
		Password: password,
		Source: "users",
	})
	if err != nil {
		LogE.Println("Application database login failed: " + err.Error())
		return err
	}

	//Init misc session
	miscDbSession = mainDbSession.Copy()
	err = miscDbSession.Login(&mgo.Credential{
		Username: username,
		Password: password,
		Source: "users",
	})
	if err != nil {
		LogE.Println("Misc database login failed: " + err.Error())
		return err
	}

	//Init reviewers session
	reviewerDbSession = mainDbSession.Copy()
	err = reviewerDbSession.Login(&mgo.Credential{
		Username: username,
		Password: password,
		Source: "users",
	})
	if err != nil {
		LogE.Println("Reviewer database login failed: " + err.Error())
		return err
	}

	return nil
}
func GetNewUserDatabase() *mgo.Database {
	s := userDbSession.Copy()
	return s.DB("users")
}
func GetNewApplicationDatabase() *mgo.Database {
	s := applicationDbSession.Copy()
	return s.DB("applications")
}
func GetNewMiscDatabase() *mgo.Database {
	s := applicationDbSession.Copy()
	return s.DB("misc")
}
func GetNewReviewerDatabase() *mgo.Database {
	s := reviewerDbSession.Copy()
	return s.DB("reviewers")
}

func initSession(){
	SessionStorage = sessions.NewCookieStore([]byte(/*NewHashString()*/"main-session-storage"))
	SessionStorage.MaxAge(86400 * 3) //3 days
}

func initStorage(){
	if !Config.IsSet("storage.serviceAccountEmail") || !Config.IsSet("storage.privateKeyPath") {
		panic(errors.New("storage.serviceAccountEmail or storage.privateKeyPath not set"))
	}

	StorageServiceAccountEmail = Config.GetString("storage.serviceAccountEmail")
	//LogD.Println("Service account: " + StorageServiceAccountEmail)

	privateKeyPath := Config.GetString("storage.privateKeyPath")
	if file,err := os.Open(privateKeyPath); err != nil {
		panic(errors.New("storage.privateKeyPath not exist"))
	}else{
		defer file.Close()
		if StoragePrivateKey,err = ioutil.ReadAll(file); err != nil {
			panic(errors.New("storage.privateKeyPath read file error"))
		}else{
			//LogD.Printf("Private key length: %d\n", len(StoragePrivateKey))
		}
	}
}

func initRecommLetter(){
	RecommLetterSubject = `NTHU A+ Recommendation Request`;
	recommLetterContentTmpl := `
		Hi,
		This is NTHU A+ management team.
		{{.ApplyUser.Name}}({{.ApplyUser.Email}}) wanted to request for your recommendation in his/her application.
		If you know about this, please visit {{.RecommUrl}} to leave your recommendation.

		Best Regards,

		NTHU A+ management team
		http://www.nthuaplus.org
	`;

	var err error
	RecommLetterTmpl,err = template.New("recommLetter").Parse(recommLetterContentTmpl)
	if err != nil {
		panic(errors.New("Error parsing recommendation letter template: " + err.Error()))
	}
}

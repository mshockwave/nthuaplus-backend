package public

import (
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"io"
	"os"
	"log"
	"io/ioutil"
	"github.com/gorilla/sessions"
)

const(
	MAIN_STORAGE_BUCKET = "nthu-a-plus-storage"

	USER_AUTH_SESSION = "user-auth"
	USER_ID_SESSION_KEY = "user_id"
)

var(

	CONFIG_FILE_NAME string = "config"
	Config	*viper.Viper

	DB_ADDRESS string = "127.0.0.1"

	mainDbSession *mgo.Session
	userDbSession *mgo.Session
	applicationDbSession *mgo.Session

	//Loggers
	LogV	*log.Logger
	LogD	*log.Logger
	LogE	*log.Logger
	LogW	*log.Logger

	//Session
	SessionStorage *sessions.CookieStore
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

func initSession(){
	SessionStorage = sessions.NewCookieStore([]byte(NewHashString()))
	SessionStorage.MaxAge(86400 * 3) //3 days
}

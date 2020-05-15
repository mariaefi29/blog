package config

import (
	"log"
	"os"

	"github.com/globalsign/mgo"
)

var (
	//SMTPEmail contains email of google smtp server
	SMTPEmail string
	//SMTPPassword contains password of google smtp server
	SMTPPassword string
	//ReCaptchaSecretCode contains secret code for recaptcha validation
	ReCaptchaSecretCode string
)

// DB instance of MongoDB
var DB *mgo.Database

// Posts are posts in a blog
var Posts *mgo.Collection

// Comments are comments to posts in a blog
var Comments *mgo.Collection

// Emails are subscription emails
var Emails *mgo.Collection

// Session is a mongo session
var Session *mgo.Session

func init() {
	//smtp server credentials
	SMTPEmail = os.Getenv("SMTP_EMAIL")
	SMTPPassword = os.Getenv("SMTP_PASSWORD")

	ReCaptchaSecretCode = os.Getenv("RECAPTCHA_SECRET")

	// get a mongo sessions
	//DB_CONNECTION_STRING = mongodb://localhost/blog (env variable)
	var err error

	DbConnectionString := os.Getenv("DB_CONNECTION_STRING")

	if DbConnectionString == "" {
		log.Println("env variable DB_CONNECTION_STRING is not defined")
	}

	Session, err = mgo.Dial(DbConnectionString)
	if err != nil {
		log.Fatal("cannot dial mongo:", err)
	}

	if err = Session.Ping(); err != nil {
		log.Fatal("cannot ping mongo:", err)
	}

	mgo.SetStats(true)

	DB = Session.DB("blog")
	// fmt.Println(DB)
	Posts = DB.C("posts")
	Comments = DB.C("comments")
	Emails = DB.C("emails")
	index := mgo.Index{
		Key:    []string{"email"},
		Unique: true,
	}
	err1 := Emails.EnsureIndex(index)
	if err != nil {
		log.Println("database error: cannot make an index on emails:", err1)
	}
	// fmt.Println("You connected to your mongo database.")
}

package config

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
var DB *mongo.Database

// Posts are posts in a blog
var Posts *mongo.Collection

// Comments are comments to posts in a blog
var Comments *mongo.Collection

// Emails are subscription emails
var Emails *mongo.Collection

// Client is a MongoDB client.
var Client *mongo.Client

func init() {
	//smtp server credentials
	SMTPEmail = os.Getenv("SMTP_EMAIL")
	SMTPPassword = os.Getenv("SMTP_PASSWORD")
	ReCaptchaSecretCode = os.Getenv("RECAPTCHA_SECRET")
	DbConnectionString := os.Getenv("DB_CONNECTION_STRING")
	if DbConnectionString == "" {
		log.Println("env variable DB_CONNECTION_STRING is not defined")
		return
	}

	ctx := context.Background()

	var err error
	Client, err = mongo.Connect(options.Client().ApplyURI(DbConnectionString))
	if err != nil {
		log.Fatal("cannot connect to mongo:", err)
	}

	if err = Client.Ping(ctx, nil); err != nil {
		log.Fatal("cannot ping mongo:", err)
	}

	DB = Client.Database("blog")
	Posts = DB.Collection("posts")
	Comments = DB.Collection("comments")
	Emails = DB.Collection("emails")
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = Emails.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Println("database error: cannot make an index on emails:", err)
	}
}

func Disconnect() {
	if Client == nil {
		return
	}

	if err := Client.Disconnect(context.Background()); err != nil {
		log.Println("database error: cannot disconnect mongo:", err)
	}
}

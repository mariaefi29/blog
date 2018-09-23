package models

import (
	"net/http"

	"github.com/globalsign/mgo/bson"
	"github.com/mariaefi29/blog/config"
	"github.com/pkg/errors"
)

//Email Struct
type Email struct {
	ID           bson.ObjectId `json:"id" bson:"_id"`
	EmailAddress string        `json:"email" bson:"email"`
}

//CreateEmail puts email address into a database
func CreateEmail(r *http.Request) (Email, error) {

	config.Session.Refresh()

	currentSession := config.Session.Copy()
	defer currentSession.Close()

	// get form values
	email := Email{}
	email.ID = bson.NewObjectId()
	email.EmailAddress = r.FormValue("email")

	// validate form values
	if email.EmailAddress == "" {
		return email, errors.New("400 bad request: all fields must be complete")
	}

	// insert values to a database
	err1 := config.Emails.Insert(email)
	if err1 != nil {
		return email, errors.Wrap(err1, "500 internal server error: CreateEmail")
	}
	return email, nil
}

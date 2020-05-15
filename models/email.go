package models

import (
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
func CreateEmail(email Email) error {
	config.Session.Refresh()
	currentSession := config.Session.Copy()
	defer currentSession.Close()

	email.ID = bson.NewObjectId()
	// insert values to a database
	if err := config.Emails.Insert(email); err != nil {
		return errors.Wrap(err, "create email")
	}

	return nil
}

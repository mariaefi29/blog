package models

import (
	"log"
	"net/http"
	"strconv"

	"github.com/haisum/recaptcha"

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
	email := Email{
		ID:           bson.NewObjectId(),
		EmailAddress: r.FormValue("email"),
	}

	noshow, err := strconv.Atoi(r.FormValue("noshow"))
	if err != nil {
		log.Println(err)
	}

	// validate form values
	if email.EmailAddress == "" {
		return Email{}, errors.New("400 bad request: all fields must be complete")
	}

	if noshow != 454 {
		return Email{}, errors.New("400 bad request: you are a bot")
	}

	re := recaptcha.R{
		Secret: config.ReCaptchaSecretCode,
	}
	recaptchaResp := r.FormValue("g-recaptcha-response")
	if !re.VerifyResponse(recaptchaResp) {
		log.Println(recaptchaResp)
		log.Println(re.Secret)
		return Email{}, errors.New("400 bad request: failed to verify recaptcha")
	}

	// insert values to a database
	if err := config.Emails.Insert(email); err != nil {
		return Email{}, errors.Wrap(err, "500 internal server error: CreateEmail")
	}

	return email, nil
}

package models

import (
	"context"
	"fmt"

	"github.com/mariaefi29/blog/config"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Email Struct
type Email struct {
	ID           bson.ObjectID `json:"id" bson:"_id"`
	EmailAddress string        `json:"email" bson:"email"`
}

// CreateEmail puts email address into a database
func CreateEmail(email Email) error {
	ctx := context.Background()

	email.ID = bson.NewObjectID()
	// insert values to a database
	if _, err := config.Emails.InsertOne(ctx, email); err != nil {
		return fmt.Errorf("create email: %w", err)
	}

	return nil
}

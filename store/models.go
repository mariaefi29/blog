package store

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Post is a blog post document.
type Post struct {
	ID            bson.ObjectID `json:"id" bson:"_id"`
	IDstr         string        `json:"idstr" bson:"idstr,omitempty"`
	Name          string        `json:"name" bson:"name"`
	Category      string        `json:"category" bson:"category"`
	CategoryEng   string        `json:"categoryeng" bson:"categoryeng"`
	CreatedAt     time.Time     `json:"created_at" bson:"created_at,omitempty"`
	Date          string        `json:"date" bson:"date"`
	Images        []string      `json:"images" bson:"images"`
	Author        string        `json:"author" bson:"author"`
	Content       []string      `json:"content" bson:"content"`
	Likes         int           `json:"likes" bson:"likes"`
	Comments      []Comment     `json:"comments" bson:"comments"`
	CommentCnt    int           `json:"comments_cnt" bson:"comments_cnt"`
	IsPopular     int           `json:"popular" bson:"popular"`
	NextPostID    bson.ObjectID `json:"next_id" bson:"next_post_id,omitempty"`
	NextPostIDstr string        `json:"next_idstr" bson:"next_post_idstr,omitempty"`
	PrevPostID    bson.ObjectID `json:"prev_id" bson:"prev_post_id,omitempty"`
	PrevPostIDstr string        `json:"prev_idstr" bson:"prev_post_idstr,omitempty"`
}

// Comment is a post comment document.
type Comment struct {
	ID          bson.ObjectID `json:"id" bson:"_id"`
	PostID      bson.ObjectID `json:"post_id" bson:"post_id"`
	Content     string        `json:"content" bson:"content" schema:"message"`
	Author      string        `json:"author" bson:"author" schema:"username"`
	Email       string        `json:"email" bson:"email" schema:"email"`
	Website     string        `json:"website" bson:"website" schema:"website"`
	CreatedAt   string        `json:"time" bson:"time"`
	ApprovedFlg int           `json:"approved_flg" bson:"approved_flg"`
}

// Email is a subscription email document.
type Email struct {
	ID           bson.ObjectID `json:"id" bson:"_id"`
	EmailAddress string        `json:"email" bson:"email"`
}

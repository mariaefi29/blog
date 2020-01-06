package models

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/mariaefi29/blog/config"
	"github.com/pkg/errors"
)

//Comment Struct
type Comment struct {
	ID          bson.ObjectId `json:"id" bson:"_id"`
	PostID      bson.ObjectId `json:"post_id" bson:"post_id"`
	Content     string        `json:"content" bson:"content" schema:"message"`
	Author      string        `json:"author" bson:"author" schema:"username"`
	Email       string        `json:"email" bson:"email" schema:"email"`
	Website     string        `json:"website" bson:"website" schema:"website"`
	CreatedAt   string        `json:"time" bson:"time"`
	ApprovedFlg int           `json:"approved_flg" bson:"approved_flg"` //pending or approved. Pending by default.
}

//CreateComment puts a comment to a post into a database
func CreateComment(r *http.Request, idstr string) (Comment, Post, error) {
	decoder.IgnoreUnknownKeys(true)

	config.Session.Refresh()
	currentSession := config.Session.Copy()
	defer currentSession.Close()

	//this method protects the website from bots
	xcode2, err := strconv.Atoi(r.FormValue("xcode2"))
	if err != nil {
		log.Println(err)
	}

	if xcode2 != 776 {
		return Comment{}, Post{}, errors.New("400 bad request: you are a bot")
	}

	post, err := OnePost(idstr)
	if err != nil {
		return Comment{}, Post{}, errors.Wrap(err, "fail to find a post to comment")
	}

	if err := r.ParseForm(); err != nil {
		return Comment{}, Post{}, errors.Wrap(err, "fail to parse a comment form")
	}

	comment := Comment{
		ID:          bson.NewObjectId(),
		PostID:      bson.ObjectIdHex(idstr),
		CreatedAt:   time.Now().Format("02.01.2006 15:04:05"),
		ApprovedFlg: 0,
	}

	err = decoder.Decode(&comment, r.PostForm)
	if err != nil {
		return Comment{}, Post{}, errors.Wrap(err, "fail to decode form into a struct")
	}

	// validate form values
	if comment.Email == "" || comment.Author == "" || comment.Content == "" {
		return Comment{},  Post{}, errors.New("400 bad request: all fields must be complete")
	}

	// insert values to a database
	err = config.Comments.Insert(comment)
	if err != nil {
		return Comment{},  Post{}, errors.Wrap(err, "Database error: fail to insert a comment")
	}

	//update a post
	post.Comments = append(post.Comments, comment)
	post.CommentCnt = 0
	for _, v := range post.Comments {
		if v.ApprovedFlg == 1 {
			post.CommentCnt++
		}
	}

	err = config.Posts.Update(bson.M{"_id": post.ID}, &post)
	if err != nil {
		return Comment{}, Post{}, errors.Wrap(err, "Database error: fail to update a post with a new comment")
	}

	return comment, post, nil
}

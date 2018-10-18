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

	// get form values
	comment := Comment{}
	post := Post{}

	//this method protects the website from bots
	xcode2, err := strconv.Atoi(r.FormValue("xcode2"))
	if err != nil {
		log.Println(err)
	}

	if xcode2 != 776 {
		return comment, post, errors.New("400 bad request: you are a bot")
	}

	post, err1 := OnePost(idstr)
	if err1 != nil {
		return comment, post, errors.Wrap(err1, "fail to find a post to comment")
	}

	err2 := r.ParseForm()
	if err2 != nil {
		return comment, post, errors.Wrap(err2, "fail to parse a comment form")
	}
	comment.ID = bson.NewObjectId()
	comment.PostID = bson.ObjectIdHex(idstr)
	currentTime := time.Now()
	comment.CreatedAt = currentTime.Format("02.01.2006 15:04:05")
	comment.ApprovedFlg = 0

	err3 := decoder.Decode(&comment, r.PostForm)
	if err3 != nil {
		return comment, post, errors.Wrap(err3, "fail to decode form into a struct")
	}

	// validate form values
	if comment.Email == "" || comment.Author == "" || comment.Content == "" {
		return comment, post, errors.New("400 bad request: all fields must be complete")
	}

	// insert values to a database
	err4 := config.Comments.Insert(comment)
	if err4 != nil {
		return comment, post, errors.Wrap(err4, "Database error: fail to insert a comment")
	}

	//update a post
	post.Comments = append(post.Comments, comment)
	post.CommentCnt = 0
	for _, v := range post.Comments {
		if v.ApprovedFlg == 1 {
			post.CommentCnt++
		}
	}

	err5 := config.Posts.Update(bson.M{"_id": post.ID}, &post)
	if err5 != nil {
		return comment, post, errors.Wrap(err5, "Database error: fail to update a post with a new comment")
	}

	return comment, post, nil
}

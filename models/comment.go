package models

import (
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
func CreateComment(comment Comment, postID string) (Post, error) {
	config.Session.Refresh()
	currentSession := config.Session.Copy()
	defer currentSession.Close()

	post, err := OnePost(postID)
	if err != nil {
		return Post{}, errors.Wrap(err, "find a post to comment")
	}

	comment.ID = bson.NewObjectId()
	comment.PostID = bson.ObjectIdHex(postID)
	comment.CreatedAt = time.Now().Format(time.RFC3339)

	// insert values to a database
	err = config.Comments.Insert(comment)
	if err != nil {
		return Post{}, errors.Wrap(err, "insert a comment into comments collections")
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
		return Post{}, errors.Wrapf(err, "update a post [%s] with a new comment", post.IDstr)
	}

	return post, nil
}

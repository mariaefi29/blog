package models

import (
	"context"
	"time"

	"github.com/mariaefi29/blog/config"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Comment Struct
type Comment struct {
	ID          bson.ObjectID `json:"id" bson:"_id"`
	PostID      bson.ObjectID `json:"post_id" bson:"post_id"`
	Content     string        `json:"content" bson:"content" schema:"message"`
	Author      string        `json:"author" bson:"author" schema:"username"`
	Email       string        `json:"email" bson:"email" schema:"email"`
	Website     string        `json:"website" bson:"website" schema:"website"`
	CreatedAt   string        `json:"time" bson:"time"`
	ApprovedFlg int           `json:"approved_flg" bson:"approved_flg"` //pending or approved. Pending by default.
}

// CreateComment puts a comment to a post into a database
func CreateComment(comment Comment, postID string) (Post, error) {
	ctx := context.Background()

	post, err := OnePost(postID)
	if err != nil {
		return Post{}, errors.Wrap(err, "find a post to comment")
	}

	comment.ID = bson.NewObjectID()
	comment.PostID = post.ID
	comment.CreatedAt = time.Now().Format(time.RFC3339)

	// insert values to a database
	if _, err := config.Comments.InsertOne(ctx, comment); err != nil {
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

	result, err := config.Posts.ReplaceOne(ctx, bson.M{"_id": post.ID}, &post)
	if err != nil {
		return Post{}, errors.Wrapf(err, "update a post [%s] with a new comment", post.IDstr)
	}
	if result.MatchedCount == 0 {
		return Post{}, errors.Errorf("update a post [%s] with a new comment: no matching post", post.IDstr)
	}

	return post, nil
}

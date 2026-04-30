package models

import (
	"context"

	"github.com/mariaefi29/blog/config"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Post Struct
type Post struct {
	ID            bson.ObjectID `json:"id" bson:"_id"`
	IDstr         string        `json:"idstr" bson:"idstr,omitempty"`
	Name          string        `json:"name" bson:"name"`
	Category      string        `json:"category" bson:"category"`
	CategoryEng   string        `json:"categoryeng" bson:"categoryeng"`
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

func reverse(s []Post) []Post {
	for i := 0; i < len(s)/2; i++ {
		j := len(s) - i - 1
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// AllPosts retrieves all posts
func AllPosts() ([]Post, error) {
	ctx := context.Background()

	posts := make([]Post, 0)
	cursor, err := config.Posts.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "find all posts")
	}
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, errors.Wrap(err, "find all posts")
	}

	reverse(posts)

	return posts, nil
}

// OnePost retrieves one post by id
func OnePost(postIDstr string) (Post, error) {
	ctx := context.Background()

	post := Post{}
	if err := config.Posts.FindOne(ctx, bson.M{"idstr": postIDstr}).Decode(&post); err != nil {
		return post, errors.Wrapf(err, "find one post [%s]", postIDstr)
	}

	return post, nil
}

// PostsByCategory retrieves posts by category
func PostsByCategory(categoryEng string) ([]Post, error) {
	ctx := context.Background()

	posts := []Post{}
	cursor, err := config.Posts.Find(ctx, bson.M{"categoryeng": categoryEng})
	if err != nil {
		return nil, errors.Wrapf(err, "find posts by category [%s]", categoryEng)
	}
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, errors.Wrapf(err, "find posts by category [%s]", categoryEng)
	}

	reverse(posts)

	return posts, nil
}

// PostLike adds one like to a post
func PostLike(post Post) (int, error) {
	ctx := context.Background()

	newLike := post.Likes + 1
	post.Likes++

	result, err := config.Posts.ReplaceOne(ctx, bson.M{"_id": post.ID}, &post)
	if err != nil {
		return 0, errors.Wrapf(err, "update post [%s] with like", post.IDstr)
	}
	if result.MatchedCount == 0 {
		return 0, errors.Errorf("update post [%s] with like: no matching post", post.IDstr)
	}

	return newLike, nil
}

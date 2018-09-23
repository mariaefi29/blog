package models

import (
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/schema"
	"github.com/mariaefi29/blog/config"
	"github.com/pkg/errors"
)

var decoder = schema.NewDecoder()

//Post Struct
type Post struct {
	ID            bson.ObjectId `json:"id" bson:"_id"`
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
	NextPostID    bson.ObjectId `json:"next_id" bson:"next_post_id,omitempty"`
	NextPostIDstr string        `json:"next_idstr" bson:"next_post_idstr,omitempty"`
	PrevPostID    bson.ObjectId `json:"prev_id" bson:"prev_post_id,omitempty"`
	PrevPostIDstr string        `json:"prev_idstr" bson:"prev_post_idstr,omitempty"`
}

func reverse(s []Post) []Post {
	for i := 0; i < len(s)/2; i++ {
		j := len(s) - i - 1
		s[i], s[j] = s[j], s[i]
	}
	return s
}

//AllPosts retrieves all posts
func AllPosts() ([]Post, error) {

	config.Session.Refresh()
	currentSession := config.Session.Copy()
	defer currentSession.Close()

	posts := []Post{}

	err := config.Posts.Find(bson.M{}).All(&posts)
	if err != nil {
		return nil, errors.Wrap(err, "Database error: AllPosts")
	}

	reverse(posts)

	return posts, nil
}

//OnePost retrieves one post by id
func OnePost(postIDstr string) (Post, error) {

	config.Session.Refresh()

	currentSession := config.Session.Copy()
	defer currentSession.Close()

	post := Post{}

	err := config.Posts.Find(bson.M{"idstr": postIDstr}).One(&post)
	if err != nil {
		return post, errors.Wrap(err, "Database error: OnePost")
	}
	return post, nil
}

//PostsByCategory retrieves posts by category
func PostsByCategory(categoryEng string) ([]Post, error) {

	config.Session.Refresh()

	currentSession := config.Session.Copy()
	defer currentSession.Close()

	posts := []Post{}

	err := config.Posts.Find(bson.M{"categoryeng": categoryEng}).All(&posts)
	if err != nil {
		return nil, errors.Wrap(err, "Database error: PostsByCategory")
	}

	reverse(posts)

	return posts, nil
}

//PostLike adds one like to a post
func PostLike(post Post) (int, error) {

	config.Session.Refresh()
	currentSession := config.Session.Copy()
	defer currentSession.Close()

	newLike := post.Likes + 1
	post.Likes++

	err := config.Posts.Update(bson.M{"_id": post.ID}, &post)
	if err != nil {
		return 0, errors.Wrap(err, "Database error: PostLike")
	}
	return newLike, nil
}

//DeletePost deletes a post from a database
func DeletePost(postID string) error {

	config.Session.Refresh()
	currentSession := config.Session.Copy()
	defer currentSession.Close()

	err := config.Posts.Remove(bson.M{"_id": bson.ObjectIdHex(postID)})
	if err != nil {
		return errors.Wrap(err, "Database error: DeletePost")
	}
	return nil
}

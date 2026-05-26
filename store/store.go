package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct {
	Client   *mongo.Client
	Database *mongo.Database
	Posts    *mongo.Collection
	Comments *mongo.Collection
	Emails   *mongo.Collection
}

func New(ctx context.Context, connectionString, databaseName string) (*Store, error) {
	if connectionString == "" {
		return nil, fmt.Errorf("DB_CONNECTION_STRING is required")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("connect to mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	database := client.Database(databaseName)
	store := &Store{
		Client:   client,
		Database: database,
		Posts:    database.Collection("posts"),
		Comments: database.Collection("comments"),
		Emails:   database.Collection("emails"),
	}

	if err := store.ensureIndexes(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	return store, nil
}

func (s *Store) Disconnect(ctx context.Context) error {
	if s == nil || s.Client == nil {
		return nil
	}

	if err := s.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("disconnect mongo: %w", err)
	}

	return nil
}

func (s *Store) ensureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	if _, err := s.Emails.Indexes().CreateOne(ctx, index); err != nil {
		return fmt.Errorf("create emails index: %w", err)
	}

	return nil
}

// AllPosts retrieves all posts.
func (s *Store) AllPosts(ctx context.Context) ([]Post, error) {
	if s == nil || s.Posts == nil {
		return nil, fmt.Errorf("posts collection is not configured")
	}

	posts := make([]Post, 0)
	cursor, err := s.Posts.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("find all posts: %w", err)
	}
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, fmt.Errorf("find all posts: %w", err)
	}

	reverse(posts)

	return posts, nil
}

// OnePost retrieves one post by id.
func (s *Store) OnePost(ctx context.Context, postIDstr string) (Post, error) {
	if s == nil || s.Posts == nil {
		return Post{}, fmt.Errorf("posts collection is not configured")
	}

	post := Post{}
	if err := s.Posts.FindOne(ctx, bson.M{"idstr": postIDstr}).Decode(&post); err != nil {
		return post, fmt.Errorf("find one post [%s]: %w", postIDstr, err)
	}

	return post, nil
}

// PostsByCategory retrieves posts by category.
func (s *Store) PostsByCategory(ctx context.Context, categoryEng string) ([]Post, error) {
	if s == nil || s.Posts == nil {
		return nil, fmt.Errorf("posts collection is not configured")
	}

	posts := []Post{}
	cursor, err := s.Posts.Find(ctx, bson.M{"categoryeng": categoryEng})
	if err != nil {
		return nil, fmt.Errorf("find posts by category [%s]: %w", categoryEng, err)
	}
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, fmt.Errorf("find posts by category [%s]: %w", categoryEng, err)
	}

	reverse(posts)

	return posts, nil
}

// PostLike adds one like to a post.
func (s *Store) PostLike(ctx context.Context, post Post) (int, error) {
	if s == nil || s.Posts == nil {
		return 0, fmt.Errorf("posts collection is not configured")
	}

	newLike := post.Likes + 1
	post.Likes++

	result, err := s.Posts.ReplaceOne(ctx, bson.M{"_id": post.ID}, &post)
	if err != nil {
		return 0, fmt.Errorf("update post [%s] with like: %w", post.IDstr, err)
	}
	if result.MatchedCount == 0 {
		return 0, fmt.Errorf("update post [%s] with like: no matching post", post.IDstr)
	}

	return newLike, nil
}

// CreateComment puts a comment to a post into a database.
func (s *Store) CreateComment(ctx context.Context, comment Comment, postID string) (Post, error) {
	if s == nil || s.Comments == nil {
		return Post{}, fmt.Errorf("comments collection is not configured")
	}
	if s.Posts == nil {
		return Post{}, fmt.Errorf("posts collection is not configured")
	}

	post, err := s.OnePost(ctx, postID)
	if err != nil {
		return Post{}, fmt.Errorf("find a post to comment: %w", err)
	}

	comment.ID = bson.NewObjectID()
	comment.PostID = post.ID
	comment.CreatedAt = time.Now().Format(time.RFC3339)

	if _, err := s.Comments.InsertOne(ctx, comment); err != nil {
		return Post{}, fmt.Errorf("insert a comment into comments collections: %w", err)
	}

	post.Comments = append(post.Comments, comment)
	post.CommentCnt = 0
	for _, v := range post.Comments {
		if v.ApprovedFlg == 1 {
			post.CommentCnt++
		}
	}

	result, err := s.Posts.ReplaceOne(ctx, bson.M{"_id": post.ID}, &post)
	if err != nil {
		return Post{}, fmt.Errorf("update a post [%s] with a new comment: %w", post.IDstr, err)
	}
	if result.MatchedCount == 0 {
		return Post{}, fmt.Errorf("update a post [%s] with a new comment: no matching post", post.IDstr)
	}

	return post, nil
}

// CreateEmail puts email address into a database.
func (s *Store) CreateEmail(ctx context.Context, email Email) error {
	if s == nil || s.Emails == nil {
		return fmt.Errorf("emails collection is not configured")
	}

	email.ID = bson.NewObjectID()
	if _, err := s.Emails.InsertOne(ctx, email); err != nil {
		return fmt.Errorf("create email: %w", err)
	}

	return nil
}

func reverse(posts []Post) []Post {
	for i := 0; i < len(posts)/2; i++ {
		j := len(posts) - i - 1
		posts[i], posts[j] = posts[j], posts[i]
	}
	return posts
}

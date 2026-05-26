package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mariaefi29/blog/config"
	"github.com/mariaefi29/blog/store"
	"github.com/stretchr/testify/require"
	"gopkg.in/gomail.v2"
)

func TestNew(t *testing.T) {
	const timeout = 5 * time.Second

	srv := New(Params{
		Config: config.Config{
			HTTP: config.HTTPConfig{
				Port:      8080,
				Timeout:   timeout,
				StaticDir: "../../public",
			},
		},
	})

	require.Equal(t, ":8080", srv.Addr)
	require.Equal(t, timeout, srv.ReadHeaderTimeout)
	require.Zero(t, srv.ReadTimeout)
	require.Zero(t, srv.WriteTimeout)
	require.Zero(t, srv.IdleTimeout)

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

const testBaseURL = "http://example.com"

func requireTestStore(t *testing.T) *store.Store {
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err, "load config")
	if cfg.Mongo.ConnectionString == "" {
		t.Skip("DB_CONNECTION_STRING is not set")
	}

	db, err := store.New(context.Background(), cfg.Mongo.ConnectionString, cfg.Mongo.Database)
	require.NoError(t, err, "connect test database")
	t.Cleanup(func() {
		require.NoError(t, db.Disconnect(context.Background()), "disconnect test database")
	})

	return db
}

func withURLParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func newTestServer(t *testing.T, db *store.Store) *server {
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err, "load config")

	tpl, err := parseTemplates(cfg.Analytics.GoogleAnalyticsMeasurementID)
	require.NoError(t, err, "parse templates")

	return &server{
		templates: tpl,
		dialer:    gomail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Email, cfg.SMTP.Password),
		config:    cfg,
		store:     db,
	}
}

// TestIndex is the simplest test: check base (/) URL
func TestIndex(t *testing.T) {
	t.Parallel()
	db := requireTestStore(t)

	srv := newTestServer(t, db)
	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", testBaseURL+"/", nil)

	srv.index(writer, req)

	require.Equal(t, http.StatusOK, writer.Code)
}

// TestShow: check (/posts/show/:id) URL
// takes out all ids of all posts from a database and checks if these requests are successful
func TestShow(t *testing.T) {
	t.Parallel()
	db := requireTestStore(t)

	//retrieves all posts from a database
	allPosts, err := db.AllPosts(context.Background())
	require.NoError(t, err, "database")

	srv := newTestServer(t, db)

	//constracts requests for each id and checks if they are successful
	for i := range allPosts {
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("GET", testBaseURL+"/posts/show/"+allPosts[i].IDstr, nil)
		req = withURLParam(req, "id", allPosts[i].IDstr)

		srv.show(writer, req)

		require.Equal(t, http.StatusOK, writer.Code)
	}
}

// TestLike: check post request to (/posts/show/:id) URL
func TestLike(t *testing.T) {
	t.Parallel()
	db := requireTestStore(t)

	var updatedPost store.Post //a modifed post after a post request

	//retrieves all posts from a database
	allPosts, err := db.AllPosts(context.Background())
	require.NoError(t, err, "database")

	srv := newTestServer(t, db)

	//contracts requests for each id and checks if they are successful
	for i := range allPosts {
		ctx := context.Background()
		restored := false
		t.Cleanup(func() {
			if !restored {
				require.NoError(t, db.ReplacePost(ctx, allPosts[i]), "database")
			}
		})

		writer := httptest.NewRecorder()
		req := httptest.NewRequest("POST", testBaseURL+"/posts/show/"+allPosts[i].IDstr, nil)
		req = withURLParam(req, "id", allPosts[i].IDstr)

		srv.like(writer, req)

		require.Equal(t, http.StatusOK, writer.Code)

		updatedPost, err = db.FindPostByID(ctx, allPosts[i].ID)
		require.NoError(t, err, "database")
		//check if the number of likes was added by one after a post request
		require.Equal(t, allPosts[i].Likes+1, updatedPost.Likes)

		//put an initial post back in the database before the post request happen
		require.NoError(t, db.ReplacePost(ctx, allPosts[i]), "database")
		restored = true

	}
}

// TestLike: check (/about) URL
func TestAbout(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, nil)
	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", testBaseURL+"/about", nil)

	srv.about(writer, req)

	require.Equal(t, http.StatusOK, writer.Code)
}

// TestContact: check get request to (/contact) URL
func TestContact(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, nil)
	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", testBaseURL+"/contact", nil)

	srv.contact(writer, req)

	require.Equal(t, http.StatusOK, writer.Code)
}

// TestCategory: check get request to (/category/:category) URL
func TestCategory(t *testing.T) {
	t.Parallel()
	db := requireTestStore(t)

	//retrieves all distinct categories from a database
	ctx := context.Background()

	categories, err := db.DistinctCategories(ctx)
	require.NoError(t, err, "database")

	srv := newTestServer(t, db)
	categoryMap := make(map[string]int64) //contains category and the amount of posts in it

	//contracts requests for each category and checks if there are working
	for i, v := range categories {
		categoryMap[v], err = db.CountPostsByCategory(ctx, v)
		require.NoError(t, err, "database")
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("GET", testBaseURL+"/category/"+categories[i], nil)
		req = withURLParam(req, "category", categories[i])

		srv.category(writer, req)

		require.Equal(t, http.StatusOK, writer.Code)

		resp := writer.Result()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		num := strings.Count(string(body), `<div class="post-snippet">`) //number of posts displayed in the categoy

		//checks if the number of posts were displayed on the page correctly
		require.Equal(t, categoryMap[v], int64(num))
	}
}

func TestComment(t *testing.T) {
	t.Parallel()
	db := requireTestStore(t)

	//retrieves all posts from a database
	allPosts, err := db.AllPosts(context.Background())
	require.NoError(t, err, "database")

	srv := newTestServer(t, db)

	for i := range allPosts {
		ctx := context.Background()
		restored := false
		removedComment := false
		t.Cleanup(func() {
			if !restored {
				require.NoError(t, db.ReplacePost(ctx, allPosts[i]), "database")
			}
			if !removedComment {
				require.NoError(t, db.DeleteCommentByEmail(ctx, "test@gmail.com"), "remove test comment")
			}
		})

		//contracts a test comment
		form := url.Values{}
		form.Add("message", "Test message")
		form.Add("username", "Test user")
		form.Add("email", "test@gmail.com")
		form.Add("website", "test.com")
		form.Add("xcode2", "776")
		testComment := strings.NewReader(form.Encode())

		writer := httptest.NewRecorder()

		req := httptest.NewRequest("POST", testBaseURL+"/posts/show/"+allPosts[i].IDstr+"/comments", testComment)
		req = withURLParam(req, "id", allPosts[i].IDstr)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv.comment(writer, req)

		require.Equal(t, http.StatusOK, writer.Code)

		//put an initial post back in the database without a test comment
		require.NoError(t, db.ReplacePost(ctx, allPosts[i]), "database")
		restored = true
		require.NoError(t, db.DeleteCommentByEmail(ctx, "test@gmail.com"), "remove test comment")
		removedComment = true
	}
}

func TestSubscribe(t *testing.T) {
	t.Parallel()
	db := requireTestStore(t)

	success := "You have successfully subscribed to blog updates!"
	fail := "You are already subscribed to blog updates!"
	writer := httptest.NewRecorder()
	writer2 := httptest.NewRecorder()
	srv := newTestServer(t, db)

	ctx := context.Background()
	removedTestEmail := false
	t.Cleanup(func() {
		if !removedTestEmail {
			require.NoError(t, db.DeleteEmailByAddress(ctx, "test@gmail.com"), "database")
		}
	})

	result, err := db.FirstEmail(ctx)
	require.NoError(t, err, "database")

	form := url.Values{}
	form.Add("email", "test@gmail.com")
	form.Add("noshow", "454")

	//subscribe by a test email
	req := httptest.NewRequest("POST", testBaseURL+"/subscribe", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	srv.subscribe(writer, req)

	require.Equal(t, http.StatusOK, writer.Code)
	resp := writer.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.NoError(t, resp.Body.Close())
	require.Equal(t, success, string(body))

	form2 := url.Values{}
	form2.Add("email", result.EmailAddress)
	form2.Add("noshow", "454")
	//subscribe by an existed email

	req2 := httptest.NewRequest("POST", testBaseURL+"/subscribe", strings.NewReader(form2.Encode()))
	req2.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	srv.subscribe(writer2, req2)

	resp2 := writer2.Result()
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)

	require.NoError(t, resp2.Body.Close())

	require.Equal(t, http.StatusOK, writer2.Code)
	require.Equal(t, fail, string(body2))

	require.NoError(t, db.DeleteEmailByAddress(ctx, "test@gmail.com"), "database")
	removedTestEmail = true
}

func TestPostDate(t *testing.T) {
	testCases := map[string]struct {
		createdAt time.Time
		fallback  string
		want      string
	}{
		"formats created at": {
			createdAt: time.Date(2026, time.May, 4, 17, 30, 0, 0, time.UTC),
			fallback:  "legacy date",
			want:      "May 4, 2026",
		},
		"falls back to legacy date": {
			createdAt: time.Time{},
			fallback:  "legacy date",
			want:      "legacy date",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.want, postDate(tt.createdAt, tt.fallback))
		})
	}
}

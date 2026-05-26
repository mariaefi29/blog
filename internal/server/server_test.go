package server

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mariaefi29/blog/config"
	"github.com/mariaefi29/blog/store"
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

	if srv.Addr != ":8080" {
		t.Fatalf("server address = %q, want %q", srv.Addr, ":8080")
	}
	if srv.ReadHeaderTimeout != timeout {
		t.Fatalf("read header timeout = %s, want %s", srv.ReadHeaderTimeout, timeout)
	}
	if srv.ReadTimeout != 0 {
		t.Fatalf("read timeout = %s, want 0", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 0 {
		t.Fatalf("write timeout = %s, want 0", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 0 {
		t.Fatalf("idle timeout = %s, want 0", srv.IdleTimeout)
	}

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d", rec.Code, http.StatusOK)
	}
}

const testBaseURL = "http://example.com"

func requireTestStore(t *testing.T) *store.Store {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Mongo.ConnectionString == "" {
		t.Skip("DB_CONNECTION_STRING is not set")
	}

	db, err := store.New(context.Background(), cfg.Mongo.ConnectionString, cfg.Mongo.Database)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Disconnect(context.Background()); err != nil {
			t.Errorf("disconnect test database: %v", err)
		}
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
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	tpl, err := parseTemplates(cfg.Analytics.GoogleAnalyticsMeasurementID)
	if err != nil {
		t.Fatalf("parse templates: %v", err)
	}

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

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}

}

// TestShow: check (/posts/show/:id) URL
// takes out all ids of all posts from a database and checks if these requests are successful
func TestShow(t *testing.T) {
	t.Parallel()
	db := requireTestStore(t)

	//retrieves all posts from a database
	allPosts, err := db.AllPosts(context.Background())
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

	srv := newTestServer(t, db)

	//constracts requests for each id and checks if they are successful
	for i := range allPosts {
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("GET", testBaseURL+"/posts/show/"+allPosts[i].IDstr, nil)
		req = withURLParam(req, "id", allPosts[i].IDstr)

		srv.show(writer, req)

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		}
	}
}

// TestLike: check post request to (/posts/show/:id) URL
func TestLike(t *testing.T) {
	t.Parallel()
	db := requireTestStore(t)

	var updatedPost store.Post //a modifed post after a post request

	//retrieves all posts from a database
	allPosts, err := db.AllPosts(context.Background())
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

	srv := newTestServer(t, db)

	//contracts requests for each id and checks if they are successful
	for i := range allPosts {
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("POST", testBaseURL+"/posts/show/"+allPosts[i].IDstr, nil)
		req = withURLParam(req, "id", allPosts[i].IDstr)

		srv.like(writer, req)

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		}

		ctx := context.Background()
		updatedPost, err = db.FindPostByID(ctx, allPosts[i].ID)
		if err != nil {
			t.Errorf("Database error is %v", err)
			continue
		}
		//check if the number of likes was added by one after a post request
		if updatedPost.Likes != allPosts[i].Likes+1 {
			t.Errorf("The likes number supposed to be %d, but got %d", allPosts[i].Likes+1, updatedPost.Likes)
		} else {
			//put an initial post back in the database before the post request happen
			if err := db.ReplacePost(ctx, allPosts[i]); err != nil {
				t.Errorf("Database error is %v", err)
			}
		}

	}
}

// TestLike: check (/about) URL
func TestAbout(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, nil)
	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", testBaseURL+"/about", nil)

	srv.about(writer, req)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}
}

// TestContact: check get request to (/contact) URL
func TestContact(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t, nil)
	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", testBaseURL+"/contact", nil)

	srv.contact(writer, req)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}

}

// TestCategory: check get request to (/category/:category) URL
func TestCategory(t *testing.T) {
	t.Parallel()
	db := requireTestStore(t)

	//retrieves all distinct categories from a database
	ctx := context.Background()

	categories, err := db.DistinctCategories(ctx)
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

	srv := newTestServer(t, db)
	categoryMap := make(map[string]int64) //contains category and the amount of posts in it

	//contracts requests for each category and checks if there are working
	for i, v := range categories {
		categoryMap[v], _ = db.CountPostsByCategory(ctx, v)
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("GET", testBaseURL+"/category/"+categories[i], nil)
		req = withURLParam(req, "category", categories[i])

		srv.category(writer, req)

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		}

		resp := writer.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		num := strings.Count(string(body), `<div class="post-snippet">`) //number of posts displayed in the categoy

		//checks if the number of posts were displayed on the page correctly
		if categoryMap[v] != int64(num) {
			t.Errorf("The number of posts in the category %v, was expected %v", num, categoryMap[v])
		}
	}
}

func TestComment(t *testing.T) {
	t.Parallel()
	db := requireTestStore(t)

	//retrieves all posts from a database
	allPosts, err := db.AllPosts(context.Background())
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

	srv := newTestServer(t, db)

	for i := range allPosts {
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

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		} else {
			ctx := context.Background()
			//put an initial post back in the database without a test comment
			if err := db.ReplacePost(ctx, allPosts[i]); err != nil {
				t.Errorf("Database error is %v", err)
			}

			if err := db.DeleteCommentByEmail(ctx, "test@gmail.com"); err != nil {
				t.Errorf("cannot remove a test comment: database error is %v", err)
			}
		}
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

	result, err := db.FirstEmail(ctx)
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

	form := url.Values{}
	form.Add("email", "test@gmail.com")
	form.Add("noshow", "454")

	//subscribe by a test email
	req := httptest.NewRequest("POST", testBaseURL+"/subscribe", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	srv.subscribe(writer, req)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}
	resp := writer.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()
	if string(body) != success {
		t.Errorf("Expected a success message: %v, but got %v", success, string(body))
	}

	form2 := url.Values{}
	form2.Add("email", result.EmailAddress)
	form2.Add("noshow", "454")
	//subscribe by an existed email

	req2 := httptest.NewRequest("POST", testBaseURL+"/subscribe", strings.NewReader(form2.Encode()))
	req2.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	srv.subscribe(writer2, req2)

	resp2 := writer2.Result()
	body2, _ := ioutil.ReadAll(resp2.Body)

	defer resp2.Body.Close()

	if writer2.Code != 200 {
		t.Errorf("Response code is %v", writer2.Code)
	}
	if string(body2) != fail {
		t.Errorf("Expected a fail message: %v, but got %v", fail, string(body2))
	}

	if err := db.DeleteEmailByAddress(ctx, "test@gmail.com"); err != nil {
		t.Errorf("Database error is %v", err)
	}

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
			if got := postDate(tt.createdAt, tt.fallback); got != tt.want {
				t.Fatalf("postDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

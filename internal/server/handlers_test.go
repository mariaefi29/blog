package server

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mariaefi29/blog/config"
	"github.com/mariaefi29/blog/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const testBaseURL = "http://example.com"

func requireTestDB(t *testing.T) {
	t.Helper()
	if config.Posts == nil {
		t.Skip("DB_CONNECTION_STRING is not set")
	}
}

func withURLParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

// TestIndex is the simplest test: check base (/) URL
func TestIndex(t *testing.T) {
	t.Parallel()
	requireTestDB(t)

	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", testBaseURL+"/", nil)

	index(writer, req)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}

}

// TestShow: check (/posts/show/:id) URL
// takes out all ids of all posts from a database and checks if these requests are successful
func TestShow(t *testing.T) {
	t.Parallel()
	requireTestDB(t)

	//retrieves all posts from a database
	allPosts, err := models.AllPosts()
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

	//constracts requests for each id and checks if they are successful
	for i := range allPosts {
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("GET", testBaseURL+"/posts/show/"+allPosts[i].IDstr, nil)
		req = withURLParam(req, "id", allPosts[i].IDstr)

		show(writer, req)

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		}
	}
}

// TestLike: check post request to (/posts/show/:id) URL
func TestLike(t *testing.T) {
	t.Parallel()
	requireTestDB(t)

	updatedPost := models.Post{} //a modifed post after a post request

	//retrieves all posts from a database
	allPosts, err := models.AllPosts()
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

	//contracts requests for each id and checks if they are successful
	for i := range allPosts {
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("POST", testBaseURL+"/posts/show/"+allPosts[i].IDstr, nil)
		req = withURLParam(req, "id", allPosts[i].IDstr)

		like(writer, req)

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		}

		ctx := context.Background()
		if err := config.Posts.FindOne(ctx, bson.M{"_id": allPosts[i].ID}).Decode(&updatedPost); err != nil {
			t.Errorf("Database error is %v", err)
			continue
		}
		//check if the number of likes was added by one after a post request
		if updatedPost.Likes != allPosts[i].Likes+1 {
			t.Errorf("The likes number supposed to be %d, but got %d", allPosts[i].Likes+1, updatedPost.Likes)
		} else {
			//put an initial post back in the database before the post request happen
			if _, err := config.Posts.ReplaceOne(ctx, bson.M{"_id": allPosts[i].ID}, &allPosts[i]); err != nil {
				t.Errorf("Database error is %v", err)
			}
		}

	}
}

// TestLike: check (/about) URL
func TestAbout(t *testing.T) {
	t.Parallel()
	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", testBaseURL+"/about", nil)

	about(writer, req)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}
}

// TestContact: check get request to (/contact) URL
func TestContact(t *testing.T) {
	t.Parallel()
	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", testBaseURL+"/contact", nil)

	contact(writer, req)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}

}

// TestCategory: check get request to (/category/:category) URL
func TestCategory(t *testing.T) {
	t.Parallel()
	requireTestDB(t)

	categories := make([]string, 0)
	//retrieves all distinct categories from a database
	ctx := context.Background()

	if err := config.Posts.Distinct(ctx, "categoryeng", bson.M{}).Decode(&categories); err != nil {
		t.Errorf("Database error is %v", err)
	}

	categoryMap := make(map[string]int64) //contains category and the amount of posts in it

	//contracts requests for each category and checks if there are working
	for i, v := range categories {
		categoryMap[v], _ = config.Posts.CountDocuments(ctx, bson.M{"categoryeng": v})
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("GET", testBaseURL+"/category/"+categories[i], nil)
		req = withURLParam(req, "category", categories[i])

		category(writer, req)

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
	requireTestDB(t)

	//retrieves all posts from a database
	allPosts, err := models.AllPosts()
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

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

		comment(writer, req)

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		} else {
			ctx := context.Background()
			//put an initial post back in the database without a test comment
			if _, err := config.Posts.ReplaceOne(ctx, bson.M{"_id": allPosts[i].ID}, &allPosts[i]); err != nil {
				t.Errorf("Database error is %v", err)
			}

			if _, err := config.Comments.DeleteOne(ctx, bson.M{"email": "test@gmail.com"}); err != nil {
				t.Errorf("cannot remove a test comment: database error is %v", err)
			}
		}
	}
}

func TestSubscribe(t *testing.T) {
	t.Parallel()
	requireTestDB(t)

	success := "Вы успешно подписаны на обновления блога!"
	fail := "Вы уже были подписаны на обновления блога!"
	writer := httptest.NewRecorder()
	writer2 := httptest.NewRecorder()

	result := models.Email{}
	ctx := context.Background()

	if err := config.Emails.FindOne(ctx, bson.M{}).Decode(&result); err != nil {
		t.Errorf("Database error is %v", err)
	}

	form := url.Values{}
	form.Add("email", "test@gmail.com")
	form.Add("noshow", "454")

	//subscribe by a test email
	req := httptest.NewRequest("POST", testBaseURL+"/subscribe", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	subscribe(writer, req)

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

	subscribe(writer2, req2)

	resp2 := writer2.Result()
	body2, _ := ioutil.ReadAll(resp2.Body)

	defer resp2.Body.Close()

	if writer2.Code != 200 {
		t.Errorf("Response code is %v", writer2.Code)
	}
	if string(body2) != fail {
		t.Errorf("Expected a fail message: %v, but got %v", fail, string(body2))
	}

	if _, err := config.Emails.DeleteOne(ctx, bson.M{"email": "test@gmail.com"}); err != nil {
		t.Errorf("Database error is %v", err)
	}

}

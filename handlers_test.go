package main

import (
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/globalsign/mgo/bson"
	"github.com/julienschmidt/httprouter"
	"github.com/mariaefi29/blog/config"
	"github.com/mariaefi29/blog/models"
)

var ts *httptest.Server
var router *httprouter.Router

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	os.Exit(code)
}

func setUp() {
	router = httprouter.New()
	router.GET("/", index)
	router.POST("/subscribe", subscribe)
	router.GET("/posts/show/:id", show)
	router.POST("/posts/show/:id", like)
	router.POST("/posts/show/:id/comments", comment)
	router.GET("/about", about)
	router.GET("/category/:category", category)
	router.GET("/contact", contact)
	router.POST("/contact", sendMessage)
	ts = httptest.NewServer(router)
	defer ts.Close()
}

// TestIndex is the simplest test: check base (/) URL
func TestIndex(t *testing.T) {

	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", ts.URL+"/", nil)

	index(writer, req, nil)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}

}

// TestShow: check (/posts/show/:id) URL
// takes out all ids of all posts from a database and checks if these requests are successful
func TestShow(t *testing.T) {

	//retrieves all posts from a database
	allPosts, err := models.AllPosts()
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

	//constracts requests for each id and checks if they are successful
	for i := range allPosts {
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("GET", ts.URL+"/posts/show/"+allPosts[i].IDstr, nil)

		ps1 := httprouter.Param{Key: "id", Value: allPosts[i].IDstr}
		ps := []httprouter.Param{ps1}

		show(writer, req, ps)

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		}
	}
}

// TestLike: check post request to (/posts/show/:id) URL
func TestLike(t *testing.T) {

	updatedPost := models.Post{} //a modifed post after a post request

	//retrieves all posts from a database
	allPosts, err := models.AllPosts()
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

	//constracts requests for each id and checks if they are successful
	for i := range allPosts {
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("POST", ts.URL+"/posts/show/"+allPosts[i].IDstr, nil)

		ps1 := httprouter.Param{Key: "id", Value: allPosts[i].IDstr}
		ps := []httprouter.Param{ps1}

		like(writer, req, ps)

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		}

		err2 := config.Posts.Find(bson.M{"_id": allPosts[i].ID}).One(&updatedPost)
		if err2 != nil {
			t.Errorf("Database error is %v", err2)
		}
		//check if the number of likes was added by one after a post request
		if updatedPost.Likes != allPosts[i].Likes+1 {
			t.Errorf("The likes number supposed to be %d, but got %d", allPosts[i].Likes+1, updatedPost.Likes)
		} else {
			//put an initial post back in the database before the post request happen
			err3 := config.Posts.Update(bson.M{"_id": allPosts[i].ID}, &allPosts[i])
			if err3 != nil {
				t.Errorf("Database error is %v", err3)
			}
		}

	}
}

// TestLike: check (/about) URL
func TestAbout(t *testing.T) {
	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", ts.URL+"/about", nil)

	about(writer, req, nil)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}
}

// TestContact: check get request to (/contact) URL
func TestContact(t *testing.T) {

	writer := httptest.NewRecorder()
	req := httptest.NewRequest("GET", ts.URL+"/contact", nil)

	about(writer, req, nil)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}

}

// TestCategory: check get request to (/category/:category) URL
func TestCategory(t *testing.T) {

	var categories []string

	//retrieves all distinct categories from a database
	err1 := config.Posts.Find(nil).Distinct("categoryeng", &categories)
	if err1 != nil {
		t.Errorf("Database error is %v", err1)
	}

	categoryMap := make(map[string]int) //contains category and the amount of posts in it

	//constracts requests for each category and checks if there are working
	for i, v := range categories {
		categoryMap[v], _ = config.Posts.Find(bson.M{"categoryeng": v}).Count()
		writer := httptest.NewRecorder()
		req := httptest.NewRequest("GET", ts.URL+"/category/"+categories[i], nil)
		ps1 := httprouter.Param{Key: "category", Value: categories[i]}
		ps := []httprouter.Param{ps1}

		category(writer, req, ps)

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		}

		resp := writer.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		num := strings.Count(string(body), `<div class="post-snippet">`) //number of posts displayed in the categoy

		//checks if the number of posts were displayed on the page correctly
		if categoryMap[v] != num {
			t.Errorf("The number of posts in the category %v, was expected %v", num, categoryMap[v])
		}
	}
}

func TestComment(t *testing.T) {

	//retrieves all posts from a database
	allPosts, err := models.AllPosts()
	if err != nil {
		t.Errorf("Database error is %v", err)
	}

	for i := range allPosts {
		//constracts a test comment
		form := url.Values{}
		form.Add("message", "Test message")
		form.Add("username", "Test user")
		form.Add("email", "test@gmail.com")
		form.Add("website", "test.com")
		form.Add("xcode2", "776")
		testComment := strings.NewReader(form.Encode())

		writer := httptest.NewRecorder()
		ps1 := httprouter.Param{Key: "id", Value: allPosts[i].IDstr}
		ps := []httprouter.Param{ps1}

		req := httptest.NewRequest("POST", ts.URL+"/posts/show/"+allPosts[i].IDstr+"/comments", testComment)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		comment(writer, req, ps)

		if writer.Code != 200 {
			t.Errorf("Response code is %v", writer.Code)
		} else {
			//put an initial post back in the database without a test comment
			err3 := config.Posts.Update(bson.M{"_id": allPosts[i].ID}, &allPosts[i])
			if err3 != nil {
				t.Errorf("Database error is %v", err3)
			}
			err4 := config.Comments.Remove(bson.M{"email": "test@gmail.com"})
			if err4 != nil {
				t.Errorf("cannot remove a test comment: database error is %v", err4)
			}
		}
	}
}

func TestSubscribe(t *testing.T) {
	success := "Вы успешно подписаны на обновления блога!"
	fail := "Вы уже были подписаны на обновления блога!"
	writer := httptest.NewRecorder()
	writer2 := httptest.NewRecorder()

	result := models.Email{}
	err1 := config.Emails.Find(nil).One(&result)
	if err1 != nil {
		t.Errorf("Database error is %v", err1)
	}

	form := url.Values{}
	form.Add("email", "test@gmail.com")
	form.Add("xcode", "776")

	//remove test email from a database
	err2 := config.Emails.Remove(bson.M{"email": "test@gmail.com"})
	if err2 != nil {
		t.Errorf("Database error is %v", err2)
	}
	//subscribe by a test email
	req := httptest.NewRequest("POST", ts.URL+"/subscribe", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	subscribe(writer, req, nil)

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
	form2.Add("xcode", "776")
	//subscribe by an existed email

	req2 := httptest.NewRequest("POST", ts.URL+"/subscribe", strings.NewReader(form2.Encode()))
	req2.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	subscribe(writer2, req2, nil)

	resp2 := writer2.Result()
	body2, _ := ioutil.ReadAll(resp2.Body)

	defer resp2.Body.Close()

	if writer2.Code != 200 {
		t.Errorf("Response code is %v", writer2.Code)
	}
	if string(body2) != fail {
		t.Errorf("Expected a fail message: %v, but got %v", fail, string(body2))
	}

	err3 := config.Emails.Remove(bson.M{"email": "test@gmail.com"})
	if err3 != nil {
		t.Errorf("Database error is %v", err2)
	}

}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/globalsign/mgo"
	"github.com/julienschmidt/httprouter"
	"github.com/mariaefi29/blog/config"
	"github.com/mariaefi29/blog/models"
	"github.com/pkg/errors"
	gomail "gopkg.in/gomail.v2"
)

var tpl *template.Template

var fm = template.FuncMap{
	"truncate": truncate,
	"incline":  commentIncline,
}

type message struct {
	name    string
	email   string
	content string
}

var d = gomail.NewDialer("smtp.gmail.com", 587, config.SMTPEmail, config.SMTPPassword)

func init() {
	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("templates/*.gohtml"))
}

func truncate(s string) string {
	var numRunes = 0
	for index := range s {
		numRunes++
		k := rune(s[index])
		if (numRunes > 150) && (k == 32) {
			return s[:index]
		}
	}
	return s
}

func commentIncline(cnt int) string {
	var s string

	if (cnt == 1) || (cnt > 20 && cnt%10 == 1) {
		s = "Комментарий"
	} else if (cnt >= 2 && cnt <= 4) || (cnt > 20 && cnt%10 >= 2 && cnt%10 <= 4) {
		s = "Комментария"
	} else {
		s = "Комментариев"
	}
	s = strconv.Itoa(cnt) + " " + s
	return s
}

func index(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	posts, error := models.AllPosts()
	if error != nil {
		http.Error(w, http.StatusText(500)+" "+error.Error(), http.StatusInternalServerError)
		return
	}

	err := tpl.ExecuteTemplate(w, "index.gohtml", posts)
	if err != nil {
		log.Fatalln(err)
	}
}

func show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadGateway)
		return
	}
	post, error := models.OnePost(id)
	if error != nil {
		http.Error(w, http.StatusText(500)+" "+error.Error(), http.StatusInternalServerError)
		fmt.Println(error)
		return
	}

	err := tpl.ExecuteTemplate(w, "show.gohtml", post)
	if err != nil {
		log.Fatalln(err)
	}
}

func about(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	err := tpl.ExecuteTemplate(w, "about.gohtml", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func contact(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	error := tpl.ExecuteTemplate(w, "contact.gohtml", nil)
	if error != nil {
		log.Fatalln(error)
	}
}

func sendMessage(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	err := req.ParseForm()
	if err != nil {
		log.Fatalln(err)
	}

	xcode3, err1 := strconv.Atoi(req.FormValue("xcode3"))
	if err1 != nil {
		log.Println(err1)
	}

	if xcode3 != 776 {
		http.Error(w, http.StatusText(400), http.StatusInternalServerError)
		log.Println("400 bad request: you are a bot")
		return
	}

	msg := &message{
		name:    req.FormValue("name"),
		email:   req.FormValue("email"),
		content: req.FormValue("message"),
	}

	m := gomail.NewMessage()
	m.SetHeader("From", msg.email)
	m.SetHeader("To", "maria.efimenko29@gmail.com")
	m.SetAddressHeader("reply-to", msg.email, msg.name)
	m.SetHeader("Subject", "Блог/контактная форма")
	m.SetBody("text/html", fmt.Sprintf("<b>Сообщение</b>: %s \n <b>От</b>: %s, %s", msg.content, msg.email, msg.name))

	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
		fmt.Fprint(w, "Произошла ошибка сервера. Попробуйте ещё раз позже.")
		return
	}

	fmt.Fprint(w, "Ваше сообщение успешно отправлено!")
}

func subscribe(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	email, error := models.CreateEmail(req)
	if error != nil && mgo.IsDup(errors.Cause(error)) {
		fmt.Fprint(w, "Вы уже были подписаны на обновления блога!")
		return
	}
	if error != nil {
		fmt.Fprint(w, "Произошла ошибка сервера. Попробуйте еще раз позже.")
		return
	}
	m := gomail.NewMessage()
	m.SetHeader("From", config.SMTPEmail)
	m.SetHeader("To", "maria.efimenko29@gmail.com")
	m.SetAddressHeader("reply-to", config.SMTPEmail, "Мария")
	m.SetHeader("Subject", "Блог/новый подписчик")
	m.SetBody("text/html", fmt.Sprintf("Поприветствуйте нового подписчика: %s.", email.EmailAddress))

	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
	}
	fmt.Fprint(w, "Вы успешно подписаны на обновления блога!")
}

func category(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	category := ps.ByName("category")
	if category == "" {
		http.Error(w, http.StatusText(400), http.StatusBadGateway)
		return
	}

	posts, error := models.PostsByCategory(category)
	if error != nil {
		http.Error(w, http.StatusText(500)+" "+error.Error(), http.StatusInternalServerError)
		return
	}

	err := tpl.ExecuteTemplate(w, "category.gohtml", posts)
	if err != nil {
		log.Fatalln(err)
	}
}

func comment(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadGateway)
		return
	}

	comment, post, error := models.CreateComment(req, id)
	if error != nil {
		log.Println(error)
		fmt.Fprint(w, "Произошла ошибка сервера. Попробуйте еще раз позже.")
		return
	}
	s := `Ваш комментарий успешно записан и проходит модерацию!`

	m := gomail.NewMessage()
	m.SetHeader("From", config.SMTPEmail)
	m.SetHeader("To", "maria.efimenko29@gmail.com")
	m.SetAddressHeader("reply-to", config.SMTPEmail, "Мария")
	m.SetHeader("Subject", "Блог/новый комментарий")
	m.SetBody("text/html", fmt.Sprintf("Пост <b>%s</b> был прокомментирован пользователем <b>%s</b>: %s.<br> Необходима модерация.", post.Name, comment.Author, comment.Content))

	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
	}
	fmt.Fprint(w, s)
}

func like(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {

	type Data struct {
		Message string `json:"message"`
		NewLike int    `json:"likes"`
	}

	var sendData Data
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadGateway)
		return
	}
	post, err := models.OnePost(id)
	if err != nil {
		log.Fatalln(err)
	}

	_, err1 := req.Cookie(ps.ByName("id"))
	if err1 != nil {
		http.SetCookie(w, &http.Cookie{
			Name:  ps.ByName("id"),
			Value: "1",
		})

		newLike, error := models.PostLike(post)
		if error != nil {
			http.Error(w, http.StatusText(406)+error.Error(), http.StatusNotAcceptable)
			log.Println(error)
			return
		}
		s := `Спасибо! Ваше мнение учтено!`
		sendData.Message = s
		sendData.NewLike = newLike
		jsonSendData, _ := json.Marshal(sendData)
		fmt.Fprint(w, string(jsonSendData))
		return
	}
	s := `Ваше мнение уже было учтено! Спасибо!`
	sendData.Message = s
	sendData.NewLike = post.Likes
	jsonSendData, _ := json.Marshal(sendData)
	fmt.Fprint(w, string(jsonSendData))
}

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
	"gopkg.in/gomail.v2"
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

var d = gomail.NewDialer("smtp.mail.ru", 465, config.SMTPEmail, config.SMTPPassword)

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
	posts, err := models.AllPosts()
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tpl.ExecuteTemplate(w, "index.gohtml", posts); err != nil {
		log.Fatalln(err)
	}
}

func show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadGateway)
		return
	}

	post, err := models.OnePost(id)
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if err := tpl.ExecuteTemplate(w, "show.gohtml", post); err != nil {
		log.Fatalln(err)
	}
}

func about(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := tpl.ExecuteTemplate(w, "about.gohtml", nil); err != nil {
		log.Fatalln(err)
	}
}

func contact(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := tpl.ExecuteTemplate(w, "contact.gohtml", nil); err != nil {
		log.Fatalln(err)
	}
}

func sendMessage(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := req.ParseForm(); err != nil {
		log.Fatalln(err)
	}

	xcode3, err := strconv.Atoi(req.FormValue("xcode3"))
	if err != nil {
		log.Println(err)
		return
	}

	if xcode3 != 776 {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		log.Println("400 bad request: you are a bot")
		return
	}

	msg := &message{
		name:    req.FormValue("name"),
		email:   req.FormValue("email"),
		content: req.FormValue("message"),
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.SMTPEmail)
	m.SetHeader("To", "maria.efimenko29@gmail.com")
	m.SetAddressHeader("reply-to", config.SMTPEmail, "Мария")
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
	email, err := models.CreateEmail(req)
	if err != nil && mgo.IsDup(errors.Cause(err)) {
		fmt.Fprint(w, "Вы уже были подписаны на обновления блога!")
		return
	}
	if err != nil {
		log.Println(err)
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

	posts, err := models.PostsByCategory(category)
	if err != nil {
		http.Error(w, http.StatusText(500)+" "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tpl.ExecuteTemplate(w, "category.gohtml", posts); err != nil {
		log.Fatalln(err)
	}
}

func comment(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadGateway)
		return
	}

	comment, post, err := models.CreateComment(req, id)
	if err != nil {
		log.Println(err)
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
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	post, err := models.OnePost(id)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = req.Cookie(ps.ByName("id"))
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:  ps.ByName("id"),
			Value: "1",
		})

		newLike, err := models.PostLike(post)
		if err != nil {
			http.Error(w, http.StatusText(500)+err.Error(), http.StatusInternalServerError)
			log.Println(err)
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

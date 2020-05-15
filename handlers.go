package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/globalsign/mgo"
	"github.com/gorilla/schema"
	"github.com/haisum/recaptcha"
	"github.com/julienschmidt/httprouter"
	"github.com/mariaefi29/blog/config"
	"github.com/mariaefi29/blog/models"
	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"
)

const (
	ServerErrorMessage           = "Произошла ошибка сервера. Попробуйте ещё раз позже."
	noShowFieldSubscribe         = 454
	noShowFieldCommentAndMessage = 776
)

type dataToSend struct {
	Message string `json:"message"`
	NewLike int    `json:"likes"`
}

var tpl *template.Template

var fm = template.FuncMap{
	"truncate": truncate,
	"incline":  commentIncline,
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

type message struct {
	name    string
	email   string
	content string
}

var d = gomail.NewDialer("smtp.mail.ru", 465, config.SMTPEmail, config.SMTPPassword)

func init() {
	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("templates/*.gohtml"))
}

func index(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	posts, err := models.AllPosts()
	if err != nil {
		http.Error(w, errors.Wrap(err, "find all posts").Error(), http.StatusInternalServerError)
		return
	}

	if err := tpl.ExecuteTemplate(w, "index.gohtml", posts); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(errors.Wrap(err, "execute template index"))
	}
}

func show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	post, err := models.OnePost(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := tpl.ExecuteTemplate(w, "show.gohtml", post); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(errors.Wrap(err, "execute template show"))
	}
}

func about(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	if err := tpl.ExecuteTemplate(w, "about.gohtml", nil); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(errors.Wrap(err, "execute template about"))
	}
}

func contact(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	if err := tpl.ExecuteTemplate(w, "contact.gohtml", nil); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(errors.Wrap(err, "execute template contact"))
	}
}

func sendMessage(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	xcode3, err := strconv.Atoi(req.FormValue("xcode3"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if xcode3 != noShowFieldCommentAndMessage {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	msg := &message{
		name:    req.FormValue("name"),
		email:   req.FormValue("email"),
		content: req.FormValue("message"),
	}

	messageToEmail := fmt.Sprintf("<b>Сообщение</b>: %s \n <b>От</b>: %s, %s", msg.content, msg.email, msg.name)
	if err := sendMessageToEmail("Блог/контактная форма", messageToEmail); err != nil {
		log.Println(errors.Wrap(err, "send new message to email"))
		_, _ = fmt.Fprint(w, ServerErrorMessage)
		return
	}

	_, _ = fmt.Fprint(w, "Ваше сообщение успешно отправлено!")
}

func subscribe(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	email := models.Email{
		EmailAddress: req.FormValue("email"),
	}

	noshow, err := strconv.Atoi(req.FormValue("noshow"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if email.EmailAddress == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if noshow != noShowFieldSubscribe {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	re := recaptcha.R{
		Secret: config.ReCaptchaSecretCode,
	}
	recaptchaResp := req.FormValue("g-recaptcha-response")
	if !re.VerifyResponse(recaptchaResp) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = models.CreateEmail(email)
	if err != nil && mgo.IsDup(errors.Cause(err)) {
		_, _ = fmt.Fprint(w, "Вы уже были подписаны на обновления блога!")
		return
	}
	if err != nil {
		log.Println(err)
		_, _ = fmt.Fprint(w, ServerErrorMessage)
		return
	}

	messageToEmail := fmt.Sprintf("Поприветствуйте нового подписчика: %s.", email.EmailAddress)
	if err := sendMessageToEmail("Блог/новый подписчик", messageToEmail); err != nil {
		log.Println(errors.Wrap(err, "send new subscriber to email"))
	}
}

func category(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	category := ps.ByName("category")
	if category == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	posts, err := models.PostsByCategory(category)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := tpl.ExecuteTemplate(w, "category.gohtml", posts); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(errors.Wrap(err, "execute template category"))
	}
}

func comment(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	xcode2, err := strconv.Atoi(req.FormValue("xcode2"))
	if err != nil {
		log.Println(err)
	}

	if xcode2 != noShowFieldCommentAndMessage {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	comment := models.Comment{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err = decoder.Decode(&comment, req.PostForm)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// validate form values
	if comment.Email == "" || comment.Author == "" || comment.Content == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	post, err := models.CreateComment(comment, id)
	if err != nil {
		_, _ = fmt.Fprint(w, ServerErrorMessage)
		log.Println(err)
		return
	}

	_, _ = fmt.Fprint(w, "Ваш комментарий успешно записан и проходит модерацию!")

	messageToEmail := constructMessageToEmail(post.Name, comment.Author, comment.Content)
	if err := sendMessageToEmail("Блог/новый комментарий", messageToEmail); err != nil {
		log.Println(errors.Wrap(err, "send comment to email"))
	}
}

func like(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	post, err := models.OnePost(id)
	if err != nil {
		http.NotFound(w, req)
	}

	_, err = req.Cookie(ps.ByName("id"))
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:  ps.ByName("id"),
			Value: "1",
		})

		newLike, err := models.PostLike(post)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err)
			return
		}

		sendData := dataToSend{
			Message: "Спасибо! Ваше мнение учтено!",
			NewLike: newLike,
		}
		jsonSendData, _ := json.Marshal(sendData)
		_, _ = fmt.Fprint(w, string(jsonSendData))
		return
	}

	sendData := dataToSend{
		Message: "Ваше мнение уже было учтено! Спасибо!",
		NewLike: post.Likes,
	}

	jsonSendData, _ := json.Marshal(sendData)
	_, _ = fmt.Fprint(w, string(jsonSendData))
}

func sendMessageToEmail(subject, message string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.SMTPEmail)
	m.SetHeader("To", "maria.efimenko29@gmail.com")
	m.SetAddressHeader("reply-to", config.SMTPEmail, "Мария")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", message)
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func constructMessageToEmail(name, author, content string) string {
	return fmt.Sprintf(
		"Пост <b>%s</b> был прокомментирован пользователем <b>%s</b>: %s.<br> Необходима модерация.",
		name, author, content,
	)
}

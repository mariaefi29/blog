package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
	"github.com/haisum/recaptcha"
	"github.com/mariaefi29/blog/config"
	"github.com/mariaefi29/blog/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
	tpl = template.Must(parseTemplates())
}

func renderTemplate(w http.ResponseWriter, name string, data any) error {
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}

func index(w http.ResponseWriter, req *http.Request) {
	posts, err := models.AllPosts()
	if err != nil {
		http.Error(w, fmt.Errorf("find all posts: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	if err := renderTemplate(w, "index.gohtml", posts); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template index: %w", err))
	}
}

func show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	post, err := models.OnePost(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := renderTemplate(w, "show.gohtml", post); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template show: %w", err))
	}
}

func about(w http.ResponseWriter, _ *http.Request) {
	if err := renderTemplate(w, "about.gohtml", nil); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template about: %w", err))
	}
}

func contact(w http.ResponseWriter, _ *http.Request) {
	if err := renderTemplate(w, "contact.gohtml", nil); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template contact: %w", err))
	}
}

func sendMessage(w http.ResponseWriter, req *http.Request) {
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
		log.Println(fmt.Errorf("send new message to email: %w", err))
		_, _ = fmt.Fprint(w, ServerErrorMessage)
		return
	}

	_, _ = fmt.Fprint(w, "Ваше сообщение успешно отправлено!")
}

func subscribe(w http.ResponseWriter, req *http.Request) {
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
	if err != nil && mongo.IsDuplicateKeyError(err) {
		_, _ = fmt.Fprint(w, "Вы уже были подписаны на обновления блога!")
		return
	}
	if err != nil {
		log.Println(err)
		_, _ = fmt.Fprint(w, ServerErrorMessage)
		return
	}

	_, _ = fmt.Fprint(w, "Вы успешно подписаны на обновления блога!")

	messageToEmail := fmt.Sprintf("Поприветствуйте нового подписчика: %s.", email.EmailAddress)
	if err := sendMessageToEmail("Блог/новый подписчик", messageToEmail); err != nil {
		log.Println(fmt.Errorf("send new subscriber to email: %w", err))
	}
}

func category(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	if category == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	posts, err := models.PostsByCategory(category)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := renderTemplate(w, "category.gohtml", posts); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template category: %w", err))
	}
}

func comment(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
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
		log.Println(fmt.Errorf("send comment to email: %w", err))
	}
}

func like(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	post, err := models.OnePost(id)
	if err != nil {
		http.NotFound(w, req)
	}

	_, err = req.Cookie(id)
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:  id,
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

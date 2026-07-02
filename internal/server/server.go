package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
	"github.com/haisum/recaptcha"
	"github.com/mariaefi29/blog/config"
	"github.com/mariaefi29/blog/store"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gopkg.in/gomail.v2"
)

const (
	serverErrorMessage           = "A server error occurred. Please try again later."
	noShowFieldSubscribe         = 454
	noShowFieldCommentAndMessage = 776
)

type server struct {
	templates *template.Template
	dialer    *gomail.Dialer
	config    config.Config
	store     *store.Store
}

type Params struct {
	Config config.Config
	Store  *store.Store
}

func New(params Params) *http.Server {
	tpl := template.Must(parseTemplates(params.Config.Analytics.GoogleAnalyticsMeasurementID))
	app := &server{
		templates: tpl,
		dialer:    gomail.NewDialer(params.Config.SMTP.Host, params.Config.SMTP.Port, params.Config.SMTP.Email, params.Config.SMTP.Password),
		config:    params.Config,
		store:     params.Store,
	}

	return &http.Server{
		Addr:              fmt.Sprintf(":%d", params.Config.HTTP.Port),
		Handler:           app.routes(params.Config.HTTP.StaticDir),
		ReadHeaderTimeout: params.Config.HTTP.Timeout,
	}
}

func (s *server) routes(staticDir string) http.Handler {
	router := chi.NewRouter()

	router.Get("/", s.index)
	router.Post("/subscribe", s.subscribe)
	router.Get("/posts/show/{id}", s.show)
	router.Post("/posts/show/{id}", s.like)
	router.Post("/posts/show/{id}/comments", s.comment)
	router.Get("/about", s.about)
	router.Get("/category/{category}", s.category)
	router.Get("/contact", s.contact)
	router.Post("/contact", s.sendMessage)
	router.Get("/impressum", s.impressum)
	router.Get("/datenschutzerklaerung", s.datenschutzerklaerung)

	if staticDir == "" {
		staticDir = "public"
	}
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	return router
}

func (s *server) renderTemplate(w http.ResponseWriter, name string, data any) error {
	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}

func (s *server) index(w http.ResponseWriter, req *http.Request) {
	posts, err := s.store.AllPosts(req.Context())
	if err != nil {
		http.Error(w, fmt.Errorf("find all posts: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	if err := s.renderTemplate(w, "index.gohtml", posts); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template index: %w", err))
	}
}

func (s *server) show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	post, err := s.store.OnePost(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := s.renderTemplate(w, "show.gohtml", post); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template show: %w", err))
	}
}

func (s *server) about(w http.ResponseWriter, _ *http.Request) {
	if err := s.renderTemplate(w, "about.gohtml", nil); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template about: %w", err))
	}
}

func (s *server) contact(w http.ResponseWriter, _ *http.Request) {
	if err := s.renderTemplate(w, "contact.gohtml", nil); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template contact: %w", err))
	}
}

func (s *server) impressum(w http.ResponseWriter, _ *http.Request) {
	if err := s.renderTemplate(w, "impressum.gohtml", nil); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template impressum: %w", err))
	}
}

func (s *server) datenschutzerklaerung(w http.ResponseWriter, _ *http.Request) {
	if err := s.renderTemplate(w, "datenschutzerklaerung.gohtml", nil); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template datenschutzerklaerung: %w", err))
	}
}

func (s *server) category(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	if category == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	posts, err := s.store.PostsByCategory(r.Context(), category)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := s.renderTemplate(w, "category.gohtml", posts); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(fmt.Errorf("execute template category: %w", err))
	}
}

func (s *server) sendMessage(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
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

	messageToEmail := fmt.Sprintf(
		"<b>Message</b>: %s \n <b>From</b>: %s, %s",
		req.FormValue("name"),
		req.FormValue("email"),
		req.FormValue("message"),
	)
	if err := s.sendMessageToEmail("Blog/contact form", messageToEmail); err != nil {
		log.Println(fmt.Errorf("send new message to email: %w", err))
		_, _ = fmt.Fprint(w, serverErrorMessage)
		return
	}

	_, _ = fmt.Fprint(w, "Your message has been sent.")
}

func (s *server) subscribe(w http.ResponseWriter, req *http.Request) {
	email := store.Email{
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
		Secret: s.config.Recaptcha.Secret,
	}
	recaptchaResp := req.FormValue("g-recaptcha-response")
	if !re.VerifyResponse(recaptchaResp) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.store.CreateEmail(req.Context(), email)
	if err != nil && mongo.IsDuplicateKeyError(err) {
		_, _ = fmt.Fprint(w, "You are already subscribed to blog updates!")
		return
	}
	if err != nil {
		log.Println(err)
		_, _ = fmt.Fprint(w, serverErrorMessage)
		return
	}

	_, _ = fmt.Fprint(w, "You have successfully subscribed to blog updates!")

	messageToEmail := fmt.Sprintf("Please welcome a new subscriber: %s.", email.EmailAddress)
	if err := s.sendMessageToEmail("Blog/new subscriber", messageToEmail); err != nil {
		log.Println(fmt.Errorf("send new subscriber to email: %w", err))
	}
}

func (s *server) comment(w http.ResponseWriter, req *http.Request) {
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

	var comment store.Comment
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

	post, err := s.store.CreateComment(req.Context(), comment, id)
	if err != nil {
		_, _ = fmt.Fprint(w, serverErrorMessage)
		log.Println(err)
		return
	}

	_, _ = fmt.Fprint(w, "Your comment has been submitted and is awaiting moderation!")

	messageToEmail := fmt.Sprintf(
		"The post <b>%s</b> received a comment from <b>%s</b>: %s.<br> Moderation is required.",
		post.Name, comment.Author, comment.Content,
	)

	err = s.sendMessageToEmail("Blog/new comment", messageToEmail)
	if err != nil {
		log.Println(fmt.Errorf("send comment to email: %w", err))
	}
}

func (s *server) sendMessageToEmail(subject, message string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.SMTP.Email)
	m.SetHeader("To", "maria.efimenko29@gmail.com")
	m.SetAddressHeader("reply-to", s.config.SMTP.Email, "Maria")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", message)
	if err := s.dialer.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

type dataToSend struct {
	Message string `json:"message"`
	NewLike int    `json:"likes"`
}

func (s *server) like(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	post, err := s.store.OnePost(req.Context(), id)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	_, err = req.Cookie(id)
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:  id,
			Value: "1",
		})

		newLike, err := s.store.PostLike(req.Context(), post)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err)
			return
		}

		sendData := dataToSend{
			Message: "Thanks! Glad you liked it",
			NewLike: newLike,
		}
		jsonSendData, _ := json.Marshal(sendData)
		_, _ = fmt.Fprint(w, string(jsonSendData))
		return
	}

	sendData := dataToSend{
		Message: "You’ve already liked this post. Thanks!",
		NewLike: post.Likes,
	}

	jsonSendData, _ := json.Marshal(sendData)
	_, _ = fmt.Fprint(w, string(jsonSendData))
}

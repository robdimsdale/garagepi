package login

import (
	"html/template"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . Handler

type Handler interface {
	LoginGET(w http.ResponseWriter, r *http.Request)
	LoginPOST(w http.ResponseWriter, r *http.Request)
	LogoutPOST(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger        lager.Logger
	templates     *template.Template
	cookieHandler *securecookie.SecureCookie
}

func NewHandler(
	logger lager.Logger,
	templates *template.Template,
	cookieHandler *securecookie.SecureCookie,
) Handler {
	return &handler{
		logger:        logger,
		templates:     templates,
		cookieHandler: cookieHandler,
	}
}

func (h handler) LoginGET(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "login", nil)
}

func (h handler) LoginPOST(w http.ResponseWriter, request *http.Request) {
	name := request.FormValue("name")
	pass := request.FormValue("password")
	if name != "" && pass != "" {
		// .. check credentials ..
		h.setSession(name, pass, w)
	}
	http.Redirect(w, request, "/", http.StatusFound)
}

func (h handler) LogoutPOST(w http.ResponseWriter, request *http.Request) {
	clearSession(w)
	http.Redirect(w, request, "/", http.StatusFound)
}

func (h handler) setSession(
	username string,
	password string,
	response http.ResponseWriter,
) {
	value := map[string]string{
		"name":     username,
		"password": password,
	}
	encoded, err := h.cookieHandler.Encode("session", value)
	if err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

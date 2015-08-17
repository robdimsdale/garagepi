package httphelper

import "net/http"

//go:generate counterfeiter . HTTPHelper

type HTTPHelper interface {
	Get(url string) (resp *http.Response, err error)
	RedirectToHomepage(w http.ResponseWriter, r *http.Request)
}

type httpHelper struct {
}

func NewHTTPHelper() HTTPHelper {
	return &httpHelper{}
}

func (h *httpHelper) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

func (h *httpHelper) RedirectToHomepage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

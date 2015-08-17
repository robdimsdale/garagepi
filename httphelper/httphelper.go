package httphelper

import "net/http"

//go:generate counterfeiter . HttpHelper

type HttpHelper interface {
	Get(url string) (resp *http.Response, err error)
	RedirectToHomepage(w http.ResponseWriter, r *http.Request)
}

type httpHelperImpl struct {
}

func NewHttpHelperImpl() HttpHelper {
	return &httpHelperImpl{}
}

func (h *httpHelperImpl) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

func (h *httpHelperImpl) RedirectToHomepage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

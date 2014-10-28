package garagepi

import "net/http"

type HttpHelper interface {
	Get(url string) (resp *http.Response, err error)
	RedirectToHomepage(w http.ResponseWriter, r *http.Request)
}

type HttpHelperImpl struct {
}

func NewHttpHelperImpl() *HttpHelperImpl {
	return &HttpHelperImpl{}
}

func (h *HttpHelperImpl) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

func (h *HttpHelperImpl) RedirectToHomepage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

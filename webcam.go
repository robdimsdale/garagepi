package garagepi

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func (e *Executor) handleWebcam(w http.ResponseWriter, r *http.Request) {
	resp, err := e.httpHelper.Get(e.webcamUrl + r.Form.Get("n"))
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error getting image: %v", err))
		if resp == nil {
			e.logger.Log("No image to return")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error closing image request: %v", err))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.Write(body)
}

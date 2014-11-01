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
			return
		}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error closing image request: %v", err))
		return
	}
	w.Write(body)
}

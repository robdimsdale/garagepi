package garagepi

import (
	"fmt"
	"net/http"
)

func (e *Executor) handleHomepage(w http.ResponseWriter, r *http.Request) {
	e.logger.Log("homepage")

	t, err := e.fsHelper.GetHomepageTemplate()
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error reading homepage template: %v", err))
		panic(err)
	}

	t.Execute(w, nil)
}

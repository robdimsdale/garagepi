package garagepi

import (
	"fmt"
	"net/http"
)

func (e *Executor) handleHomepage(w http.ResponseWriter, r *http.Request) {
	t, err := e.fsHelper.GetHomepageTemplate()
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error reading homepage template: %v", err))
		panic(err)
	}

	ls, err := e.discoverLightState()
	if err != nil {
		e.logger.Log("Error reading light state - rendering homepage without light controls")
	}

	t.Execute(w, ls)
}

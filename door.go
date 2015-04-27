package garagepi

import (
	"fmt"
	"net/http"
)

func (e Executor) handleDoorToggle(w http.ResponseWriter, r *http.Request) {
	e.logger.Log("Toggling door")
	err := e.gpio.WriteHigh(e.gpioDoorPin)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error toggling door (skipping sleep and further executions): %v", err))
		w.Write([]byte("error - door not toggled"))
		return
	} else {
		e.osHelper.Sleep(SleepTime)

		err := e.gpio.WriteLow(e.gpioDoorPin)
		if err != nil {
			e.logger.Log(fmt.Sprintf("Error toggling door: %v", err))
		}

		e.logger.Log("door toggled")
		w.Write([]byte("door toggled"))
		return
	}
}

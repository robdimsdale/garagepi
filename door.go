package garagepi

import (
	"fmt"
	"net/http"
)

func (e Executor) handleDoorToggle(w http.ResponseWriter, r *http.Request) {
	args := []string{GpioWriteCommand, tostr(e.gpioDoorPin), GpioHighState}
	e.logger.Log("Toggling door")
	_, err := e.executeCommand(e.gpioExecutable, args...)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error toggling door (skipping sleep and further executions): %v", err))
		w.Write([]byte("error - door not toggled"))
		return
	} else {
		e.osHelper.Sleep(SleepTime)
		e.executeCommand(e.gpioExecutable, GpioWriteCommand, tostr(e.gpioDoorPin), GpioLowState)
		e.logger.Log("door toggled")
		w.Write([]byte("door toggled"))
		return
	}
}

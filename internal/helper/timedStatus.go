package helper

import (
	"time"

	"github.com/rivo/tview"
)

func DisplayTimedStatus(app *tview.Application, message string, duration time.Duration, statusBar *tview.TextView, defaultBarText string) {
	statusBar.SetText(message)
	go func() {
		time.Sleep(duration)
		app.QueueUpdateDraw(func() {
			statusBar.SetText(defaultBarText)
		})
	}()
}

package helper

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func StatusToggle(tuiApp *tview.Application) *tview.TextView {
	textView := tview.NewTextView()

	textView.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorDarkBlue)

	textView.SetChangedFunc(func() {
		tuiApp.Draw()
	})
	textView.SetFocusFunc(func() {
		textView.SetBackgroundColor(tcell.ColorGray)
	})
	textView.SetBlurFunc(func() {
		textView.SetBackgroundColor(tcell.ColorDarkBlue)
	})

	return textView
}

func StatusBar() *tview.TextView {
	statusBar := tview.NewTextView()
	statusBar.SetTextAlign(tview.AlignLeft).
		SetTextColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorDarkGray)
	return statusBar
}

func UpdateToggleText(state string, toggleStatus *tview.TextView) {
	if state == "running" {
		toggleStatus.SetText("[::b][green][ ACTIVE ][-:-:-] [white]INACTIVE")
	} else if state == "exited" {
		toggleStatus.SetText("[white]ACTIVE [::b][red] [ INACTIVE ][-:-:-]")
	} else {
		toggleStatus.SetText("None")
	}
}

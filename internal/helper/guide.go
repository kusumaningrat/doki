package helper

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func ExitGuide() *tview.TextView {
	return tview.NewTextView().
		SetText("Use arrow keys to navigate. Press 'q', ESC or Ctrl + C to quit.").
		SetTextColor(tcell.ColorGreen).
		SetTextAlign(tview.AlignCenter)
}

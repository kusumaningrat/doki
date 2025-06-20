package helper

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func ContainerGuide() *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetText(("[white][::u]^Q[::-] Quit        [white][::u]^M[::-] Menu     [white][::u]Esc[::-] Back\n" +
			"[white][::u]^R[::-] Restart Container  [white][::u]^S[::-] Start Container   [white][::u]^X[::-] Remove Container\n" +
			"[white][::u]^A[::-] Reload Table  [white][::u]^I[::-] Inspect Container")).
		SetTextColor(tcell.ColorWhite).
		SetTextAlign(tview.AlignCenter)
}

func ImageGuide() *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetText(("[white][::u]^Q[::-] Quit        [white][::u]^M[::-] Menu     [white][::u]Esc[::-] Back\n" +
			"[white][::u]^R[::-] Remove Image  [white][::u]^S[::-] Inspect Image")).
		SetTextColor(tcell.ColorWhite).
		SetTextAlign(tview.AlignCenter)
}

func VolumeGuide() *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetText(("[white][::u]^Q[::-] Quit        [white][::u]^M[::-] Menu     [white][::u]Esc[::-] Back\n" +
			"[white][::u]^R[::-] Remove Volume  [white][::u]^S[::-] Inspect Volume")).
		SetTextColor(tcell.ColorWhite).
		SetTextAlign(tview.AlignCenter)
}

func NetworkGuide() *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetText(("[white][::u]^Q[::-] Quit        [white][::u]^M[::-] Menu     [white][::u]Esc[::-] Back\n" +
			"[white][::u]^R[::-] Remove Network  [white][::u]^S[::-] Inspect Network")).
		SetTextColor(tcell.ColorWhite).
		SetTextAlign(tview.AlignCenter)
}

package inspect

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InspectModal represents the inspect modal view
type InspectModal struct {
	*tview.Flex
	textView *tview.TextView
}

// GetContentPrimitive returns the primary focusable primitive (TextView)
func (m *InspectModal) GetContentPrimitive() tview.Primitive {
	return m.textView
}

// CreateInspectModal creates and returns the inspect modal's content
func CreateInspectModal(
	app *tview.Application,
	pages *tview.Pages,
	table *tview.Table,
	rawJSON string,
	title string,
	displayStatus func(message string, duration time.Duration),
) *InspectModal {
	textView := tview.NewTextView().
		SetScrollable(true).
		SetWordWrap(false).
		SetText(rawJSON)

	textView.SetBackgroundColor(tcell.ColorBlack)
	textView.SetTextColor(tcell.ColorWhite)
	textView.SetBorder(true).
		SetBorderColor(tcell.ColorGreen)
	textView.SetTitle(fmt.Sprintf(" Inspecting %s ", title)).
		SetTitleAlign(tview.AlignCenter)

	closeButton := tview.NewButton("Close").SetSelectedFunc(func() {
		pages.RemovePage("inspect_view")
		app.SetFocus(table)
	})

	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(closeButton, 15, 0, true).
		AddItem(nil, 0, 1, false)

	inspectContentFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, true).
		AddItem(buttonContainer, 1, 0, false)

	centeredModal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(inspectContentFlex, 0, 3, true).
			AddItem(nil, 0, 1, false),
			0, 3, false).
		AddItem(nil, 0, 1, false)

	return &InspectModal{
		Flex:     centeredModal,
		textView: textView,
	}
}

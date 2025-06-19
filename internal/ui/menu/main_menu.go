package menu

import (
	"docker-tui/internal/helper"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	PageMainMenu      = "main_menu"
	PageContainerList = "container_list"
	PageInspectView   = "inspect_view"
)

func CreateMainMenu(
	app *tview.Application,
	p *tview.Pages,
	containerTable *tview.Table, // Only containerTable needed now
	displayTimedStatus func(message string, duration time.Duration,
	)) *tview.Flex {

	headerTextView := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Docker Management - CLI Based").
		SetTextColor(tcell.ColorGreen).
		SetMaxLines(1)

	createCenteredButton := func(text string, selectedFunc func()) *tview.Flex {
		button := tview.NewButton(text).SetSelectedFunc(func() {
			selectedFunc() // This is your original p.SwitchToPage(PageContainerList) etc.
		})

		buttonWidth := len(text) + 4
		if buttonWidth < 20 { // Ensure a minimum reasonable width for consistency, adjust as needed
			buttonWidth = 20
		}

		// Create a horizontal Flex container to center this single button
		return tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(tview.NewBox(), 0, 1, false).  // Left flexible spacer
			AddItem(button, buttonWidth, 0, true). // The button itself with a fixed width, focusable
			AddItem(tview.NewBox(), 0, 1, false)   // Right flexible spacer
	}

	containersButton := createCenteredButton("Containers", func() {
		p.SwitchToPage(PageContainerList)
		app.SetFocus(containerTable)
	})

	imagesButton := createCenteredButton("Images", func() {
		helper.DisplayTimedStatus(app, "Images page (not implemented yet)", 2*time.Second, nil, "")
	})

	volumesButton := createCenteredButton("Volumes", func() {
		helper.DisplayTimedStatus(app, "Volumes page (not implemented yet)", 2*time.Second, nil, "")
	})

	quitButton := createCenteredButton("Quit", func() {
		app.Stop()
	})

	menuFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 0, 1, false). // Top Spacer (flexible height)
		AddItem(headerTextView, 1, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(containersButton, 1, 0, true).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(imagesButton, 1, 0, true).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(volumesButton, 1, 0, true).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(quitButton, 1, 0, true).
		AddItem(tview.NewBox(), 0, 1, false) // Bottom Spacer

	// Wrap the menuFlex in another Flex for horizontal centering
	centeredMenu := tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(menuFlex, 0, 1, true).
		AddItem(tview.NewBox(), 0, 1, false)

	// Set a title for the centered menu (optional)
	centeredMenu.SetBorder(true).SetTitle(" Main Menu ").SetTitleAlign(tview.AlignCenter)

	return centeredMenu
}

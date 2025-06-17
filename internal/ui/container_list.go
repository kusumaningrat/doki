package ui

import "github.com/rivo/tview"

func CreateContainerListPage(
	app *tview.Application,
	table *tview.Table,
	statusToggle *tview.TextView,
	statusBar *tview.TextView,
	exitGuide *tview.TextView) *tview.Flex {
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(statusToggle, 1, 0, false).
		AddItem(table, 0, 1, true). // Table is the primary focus of this view
		AddItem(statusBar, 1, 0, false).
		AddItem(exitGuide, 1, 0, false) // This guide might instruct to press 'm' for menu
	return layout
}

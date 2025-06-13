package handler

import (
	"context"
	"docker-tui/internal/app"

	"fmt"
	"os"

	"docker-tui/internal/helper"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Container struct {
	ID      string
	Image   string
	Command string
	Created string
	Status  string
	Ports   string
	Names   string
}

func RunCLI(app *app.ContainerUseCase) {
	tuiApp := tview.NewApplication()

	var currentFilteredState string = "running"

	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, true)
	table.SetFixed(1, 0)

	headers := []string{
		"CONTAINER ID", "IMAGE", "COMMAND", "CREATED", "STATUS", "PORTS", "NAMES",
	}

	table.SetTitle("Docker Container - CLI Based").SetTitleAlign(tview.AlignCenter)

	for col, header := range headers {
		table.SetCell(0, col,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
	}

	// containers, err := app.ListAllContainers(context.Background())
	// containers, err := app.ListContainersByState(context.Background(), currentFilteredState)
	// if err != nil {
	// 	fmt.Printf("Failed to fetch containers data: %v\n", err)
	// }

	statusToggle := interface{}(tview.NewTextView()).(*tview.TextView)
	statusToggle.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorDarkBlue)

	statusToggle.SetChangedFunc(func() {
		tuiApp.Draw()
	})
	statusToggle.SetFocusFunc(func() {
		statusToggle.SetBackgroundColor(tcell.ColorGray)
	})
	statusToggle.SetBlurFunc(func() {
		statusToggle.SetBackgroundColor(tcell.ColorDarkBlue)
	})

	updateToggleText := func() {
		if currentFilteredState == "running" {
			statusToggle.SetText("[::b][green][ ACTIVE ][-:-:-] [white]INACTIVE")
		} else if currentFilteredState == "exited" {
			statusToggle.SetText("[white]ACTIVE [::b][red] [ INACTIVE ][-:-:-]")
		} else {
			statusToggle.SetText("dada")
		}
	}

	// Initialize UI
	updateToggleText()

	statusBar := interface{}(tview.NewTextView()).(*tview.TextView)
	statusBar.SetTextAlign(tview.AlignLeft).
		SetTextColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorDarkGray)

	refreshTable := func(state string) {
		currentFilteredState = state
		updateToggleText()

		containers, err := app.ListContainersByState(context.Background(), currentFilteredState)
		if err != nil {
			statusBar.SetText(fmt.Sprintf("Error: %v", err))
			return
		}

		// Clear old rows
		for row := table.GetRowCount() - 1; row >= 1; row-- {
			table.RemoveRow(row)
		}

		for rowNum, container := range containers {
			rowIdx := rowNum + 1
			table.SetCell(rowIdx, 0, tview.NewTableCell(container.ID[:12]).SetAlign(tview.AlignLeft).SetReference(&containers[rowNum]))
			table.SetCell(rowIdx, 1, tview.NewTableCell(container.Image).SetAlign(tview.AlignLeft))
			table.SetCell(rowIdx, 2, tview.NewTableCell(container.Command).SetAlign(tview.AlignLeft))
			table.SetCell(rowIdx, 3, tview.NewTableCell(container.Created).SetAlign(tview.AlignLeft))
			table.SetCell(rowIdx, 4, tview.NewTableCell(container.Status).SetAlign(tview.AlignLeft))
			table.SetCell(rowIdx, 5, tview.NewTableCell(helper.FormatContainerPorts(container.Ports)).SetAlign(tview.AlignLeft))
			table.SetCell(rowIdx, 6, tview.NewTableCell(container.Name).SetAlign(tview.AlignLeft))
		}
	}

	exitGuide := tview.NewTextView().
		SetText("Use arrow keys to navigate. Press 'q', ESC or Ctrl + C to quit.").
		SetTextColor(tcell.ColorGreen).
		SetTextAlign(tview.AlignCenter)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(statusToggle, 1, 0, true).
		AddItem(table, 0, 1, true).
		AddItem(statusBar, 1, 0, false).
		AddItem(exitGuide, 1, 0, true)

	refreshTable(currentFilteredState)

	tuiApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if tuiApp.GetFocus() == statusToggle {
				tuiApp.SetFocus(table)
				statusBar.SetText("Focus: table")
			} else {
				tuiApp.SetFocus(statusToggle)
				statusBar.SetText("Focus: statusToggle")
			}
			return nil
		case tcell.KeyEnter:
			if tuiApp.GetFocus() == statusToggle {
				if currentFilteredState == "running" {
					refreshTable("exited")
				} else {
					refreshTable("running")
				}
				statusBar.SetText(fmt.Sprintf("Switched to: %s", currentFilteredState))
				return nil
			}
		case tcell.KeyEscape, tcell.KeyCtrlC:
			tuiApp.Stop()
			return nil
		default:
			if event.Rune() == 'q' {
				tuiApp.Stop()
				return nil
			}
		}
		return event
	})

	tuiApp.SetRoot(flex, true).EnableMouse(true).SetFocus(table)

	if err := tuiApp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}

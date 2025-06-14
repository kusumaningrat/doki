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
	Name    string
}

func RunCLI(app *app.ContainerUseCase) {
	tuiApp := tview.NewApplication()

	var currentFilteredState string = "running"

	table := helper.TableFormat()

	statusToggle := helper.StatusToggle(tuiApp)

	// Initialize UI
	helper.UpdateToggleText(currentFilteredState, statusToggle)

	statusBar := helper.StatusBar()

	refreshTable := func(state string) {
		currentFilteredState = state
		helper.RefreshTable(context.Background(), state, statusToggle, table, statusBar, app)
	}

	exitGuide := helper.ExitGuide()

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

package handler

import (
	"context"
	"docker-tui/internal/app"
	"docker-tui/internal/domain"
	"strings"
	"time"

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

type ContainerUseCases = app.ContainerUseCases

var autoRefreshInterval = 3 * time.Second
var autoRefreshEnabled = false
var stopAutoRefresh chan struct{}

func RunCLI(usecase *ContainerUseCases) {
	tuiApp := tview.NewApplication()

	var currentFilteredState string = "running"
	const defaultBarText = ""

	table := helper.TableFormat()

	statusToggle := helper.StatusToggle(tuiApp)

	// Initialize UI
	helper.UpdateToggleText(currentFilteredState, statusToggle)

	statusBar := helper.StatusBar()

	stopAutoRefresh = make(chan struct{})

	displayTimedStatus := func(message string, duration time.Duration) {
		statusBar.SetText(message) // Set the new message
		go func() {                // Run the timer in a new goroutine to avoid blocking the UI
			time.Sleep(duration)
			tuiApp.QueueUpdateDraw(func() { // Queue the update back on the main UI thread
				statusBar.SetText(defaultBarText) // Revert to the default guide text
			})
		}()
	}

	refreshTable := func(state string) {
		currentFilteredState = state
		containers, err := usecase.Query.ListContainersByState(context.Background(), state)
		if err != nil {
			displayTimedStatus(fmt.Sprintf("Error fetching containers: %v", err), 3*time.Second)
			return
		}
		helper.PopulateContainerTableUI(table, containers)
		helper.UpdateToggleText(state, statusToggle)
	}

	exitGuide := helper.ExitGuide()

	startAutoRefresh := func() {
		if autoRefreshEnabled { // Prevent starting if already enabled
			return
		}
		autoRefreshEnabled = true
		go func() {
			ticker := time.NewTicker(autoRefreshInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					refreshTable(currentFilteredState)
				case <-stopAutoRefresh: // Listen for stop signal
					return
				}
			}
		}()
		displayTimedStatus(fmt.Sprintf("Auto-refresh enabled (every %s). Press 'a' to disable.", autoRefreshInterval), 3*time.Second)
	}

	stopAutoRefreshFunc := func() {
		if !autoRefreshEnabled {
			return
		}
		autoRefreshEnabled = false
		close(stopAutoRefresh)                // Send stop signal
		stopAutoRefresh = make(chan struct{}) // Re-create channel for next start (important!)
		displayTimedStatus("Auto-refresh disabled. Press 'a' to enable.", 3*time.Second)
	}

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(statusToggle, 1, 0, true).
		AddItem(table, 0, 1, true).
		AddItem(statusBar, 1, 0, false).
		AddItem(exitGuide, 1, 0, true)

	refreshTable(currentFilteredState)
	startAutoRefresh() // Start auto refresh when initiating the app

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

			if event.Rune() == 's' {
				if tuiApp.GetFocus() == table {
					row, _ := table.GetSelection()
					if row < 1 {
						displayTimedStatus("No container selected (header row).", 2*time.Second)
						return nil
					}
					cell := table.GetCell(row, 0)
					if cell == nil {
						displayTimedStatus("Error: Selected cell is nil.", 3*time.Second)
						return nil
					}
					containerRef := cell.GetReference()
					if containerRef == nil {
						displayTimedStatus("No container reference found.", 3*time.Second)
						return nil
					}
					container := containerRef.(*domain.Container)
					err := usecase.Control.StartContainer(context.Background(), container.ID)
					if err != nil {
						displayTimedStatus(fmt.Sprintf("Failed to start: %v", err), 5*time.Second)
					} else {
						displayTimedStatus(fmt.Sprintf("Container %s started. Switched to Active view.", container.ID[:12]), 3*time.Second)
						currentFilteredState = "running"
						helper.UpdateToggleText(currentFilteredState, statusToggle)

						refreshTable(currentFilteredState)
						statusBar.SetText(fmt.Sprintf("Container %s started. Switched to Active view.", container.ID[:12]))
					}
					return nil
				}
			}
			if event.Rune() == 'x' {
				if tuiApp.GetFocus() == table {
					row, _ := table.GetSelection()
					if row < 1 {
						displayTimedStatus("No container selected (header row).", 2*time.Second)
						return nil
					}
					cell := table.GetCell(row, 0)
					if cell == nil {
						displayTimedStatus("Error: Selected cell is nil.", 3*time.Second)
						return nil
					}
					containerRef := cell.GetReference()
					if containerRef == nil {
						displayTimedStatus("No container reference found.", 3*time.Second)
						return nil
					}
					container := containerRef.(*domain.Container)
					err := usecase.Control.StopContainer(context.Background(), container.ID)
					if err != nil {
						displayTimedStatus(fmt.Sprintf("Failed to stop: %v", err), 5*time.Second)
					} else {
						displayTimedStatus(fmt.Sprintf("Container %s stopped.", container.ID[:12]), 3*time.Second)
						helper.UpdateToggleText(currentFilteredState, statusToggle)
						refreshTable(currentFilteredState)

						// refreshTable(currentFilteredState)
						// statusBar.SetText(fmt.Sprintf("Container %s started. Switched to Active view.", container.ID[:12]))
					}
					return nil
				}
			}
			if event.Rune() == 'r' {
				if tuiApp.GetFocus() == table {
					row, _ := table.GetSelection()
					if row < 1 {
						displayTimedStatus("No container selected (header row).", 2*time.Second)
						return nil
					}
					cell := table.GetCell(row, 0)
					if cell == nil {
						displayTimedStatus("Error: Selected cell is nil.", 3*time.Second)
						return nil
					}
					containerRef := cell.GetReference()
					if containerRef == nil {
						displayTimedStatus("No container reference found.", 3*time.Second)
						return nil
					}
					container := containerRef.(*domain.Container)
					err := usecase.Control.RestartContainer(context.Background(), container.ID)
					if err != nil {
						displayTimedStatus(fmt.Sprintf("Failed to restart: %v", err), 5*time.Second)
					} else {
						displayTimedStatus(fmt.Sprintf("Container %s restarted.", container.ID[:12]), 3*time.Second)
						helper.UpdateToggleText(currentFilteredState, statusToggle)
						refreshTable(currentFilteredState)

					}
					return nil
				}
			}
			if event.Rune() == 'd' {
				if tuiApp.GetFocus() == table {
					row, _ := table.GetSelection()
					if row < 1 {
						displayTimedStatus("No container selected (header row).", 2*time.Second)
						return nil
					}
					cell := table.GetCell(row, 0)
					if cell == nil {
						displayTimedStatus("Error: Selected cell is nil.", 3*time.Second)
						return nil
					}
					containerRef := cell.GetReference()
					if containerRef == nil {
						displayTimedStatus("No container reference found.", 3*time.Second)
						return nil
					}
					container := containerRef.(*domain.Container)
					forceRemove := false
					if strings.HasPrefix(container.Status, "Up") {
						forceRemove = true
						displayTimedStatus(fmt.Sprintf("Attempting to force remove running container %s...", container.ID[:12]), 2*time.Second)
					} else {
						displayTimedStatus(fmt.Sprintf("Attempting to remove stopped container %s...", container.ID[:12]), 2*time.Second)
					}
					err := usecase.Control.RemoveContainer(context.Background(), container.ID, forceRemove)
					if err != nil {
						displayTimedStatus(fmt.Sprintf("Failed to remove: %v", err), 5*time.Second)
					} else {
						displayTimedStatus(fmt.Sprintf("Container %s removed.", container.ID[:12]), 3*time.Second)
						helper.UpdateToggleText(currentFilteredState, statusToggle)
						refreshTable(currentFilteredState)

					}
					return nil
				}
			}
			if event.Rune() == 'a' { // 'a' for auto-refresh toggle
				if autoRefreshEnabled {
					stopAutoRefreshFunc()
				} else {
					startAutoRefresh()
				}
				return nil // Consume the 'a' event
			}
		}
		return event
	})

	tuiApp.SetRoot(flex, true).EnableMouse(true).SetFocus(table)

	if err := tuiApp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}

	defer stopAutoRefreshFunc()
}

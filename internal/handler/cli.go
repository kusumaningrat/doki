package handler

import (
	"bytes"
	"context"
	"docker-tui/internal/app"
	"docker-tui/internal/domain"
	"encoding/json"
	"strings"
	"time"

	"fmt"
	"os"

	"docker-tui/internal/helper"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ContainerUseCases = app.ContainerUseCases

// Global variables for tview components (accessible across RunCLI and its closures)
var tuiApp *tview.Application
var pages *tview.Pages
var table *tview.Table
var mainAppLayout *tview.Flex // Your main app layout containing table, status bar, etc.

var autoRefreshInterval = 3 * time.Second
var autoRefreshEnabled = false
var stopAutoRefresh chan struct{}

func RunCLI(usecase *ContainerUseCases) {
	tuiApp = tview.NewApplication()

	var currentFilteredState string = "running"
	const defaultBarText = ""

	table = helper.TableFormat() // Initialize the table

	statusToggle := helper.StatusToggle(tuiApp)
	statusBar := helper.StatusBar()
	exitGuide := helper.ExitGuide()

	// Initialize UI elements
	helper.UpdateToggleText(currentFilteredState, statusToggle)

	stopAutoRefresh = make(chan struct{})

	// Function to display timed status messages in the status bar
	displayTimedStatus := func(message string, duration time.Duration) {
		statusBar.SetText(message)
		go func() {
			time.Sleep(duration)
			tuiApp.QueueUpdateDraw(func() {
				statusBar.SetText(defaultBarText)
			})
		}()
	}

	// Function to refresh the container table
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

	// Functions for auto-refresh
	startAutoRefresh := func() {
		if autoRefreshEnabled {
			return
		}
		autoRefreshEnabled = true
		go func() {
			ticker := time.NewTicker(autoRefreshInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					tuiApp.QueueUpdateDraw(func() {
						refreshTable(currentFilteredState)
					})
				case <-stopAutoRefresh:
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
		close(stopAutoRefresh)
		stopAutoRefresh = make(chan struct{})
		displayTimedStatus("Auto-refresh disabled. Press 'a' to enable.", 3*time.Second)
	}

	// --- Main Application Layout and Pages Setup ---
	mainAppLayout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(statusToggle, 1, 0, false).
		AddItem(table, 0, 1, true). // Table is the primary focus of the main view
		AddItem(statusBar, 1, 0, false).
		AddItem(exitGuide, 1, 0, false)

	pages = tview.NewPages().
		AddPage("main", mainAppLayout, true, true) // Add mainAppLayout as the initial page

	// Initial refresh and start auto-refresh
	refreshTable(currentFilteredState)
	startAutoRefresh()

	tuiApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentPageName, _ := pages.GetFrontPage()

		if currentPageName == "inspect_view" {
			return event
		}

		// Main application input capture logic (only active when "main" page is front)
		if currentPageName == "main" {
			switch event.Key() {
			case tcell.KeyTab:
				if tuiApp.GetFocus() == statusToggle {
					tuiApp.SetFocus(table)
					displayTimedStatus("Focus: table", 1*time.Second)
				} else {
					tuiApp.SetFocus(statusToggle)
					displayTimedStatus("Focus: statusToggle", 1*time.Second)
				}
				return nil
			case tcell.KeyEnter:
				if tuiApp.GetFocus() == statusToggle {
					if currentFilteredState == "running" {
						refreshTable("exited")
					} else {
						refreshTable("running")
					}
					displayTimedStatus(fmt.Sprintf("Switched to: %s", currentFilteredState), 2*time.Second)
					return nil
				}
			case tcell.KeyEscape, tcell.KeyCtrlC: // Global app exit
				tuiApp.Stop()
				return nil
			default:
				if event.Rune() == 'q' { // 'q' to quit (main app)
					tuiApp.Stop()
					return nil
				}

				if tuiApp.GetFocus() == table {
					row, _ := table.GetSelection()
					if row < 1 { // Header row check
						displayTimedStatus("No container selected (header row).", 2*time.Second)
						return nil
					}
					cell := table.GetCell(row, 0)
					if cell == nil || cell.GetReference() == nil {
						displayTimedStatus("Error: No container reference found.", 3*time.Second)
						return nil
					}
					container := cell.GetReference().(*domain.Container)

					switch event.Rune() {
					case 's': // Start container
						err := usecase.Control.StartContainer(context.Background(), container.ID)
						if err != nil {
							displayTimedStatus(fmt.Sprintf("Failed to start: %v", err), 5*time.Second)
						} else {
							displayTimedStatus(fmt.Sprintf("Container %s started. Switched to Active view.", container.ID[:12]), 3*time.Second)
							currentFilteredState = "running"
							helper.UpdateToggleText(currentFilteredState, statusToggle)
							refreshTable(currentFilteredState)
						}
						return nil
					case 'x': // Stop container
						err := usecase.Control.StopContainer(context.Background(), container.ID)
						if err != nil {
							displayTimedStatus(fmt.Sprintf("Failed to stop: %v", err), 5*time.Second)
						} else {
							displayTimedStatus(fmt.Sprintf("Container %s stopped.", container.ID[:12]), 3*time.Second)
							helper.UpdateToggleText(currentFilteredState, statusToggle)
							refreshTable(currentFilteredState)
						}
						return nil
					case 'r': // Restart container
						err := usecase.Control.RestartContainer(context.Background(), container.ID)
						if err != nil {
							displayTimedStatus(fmt.Sprintf("Failed to restart: %v", err), 5*time.Second)
						} else {
							displayTimedStatus(fmt.Sprintf("Container %s restarted.", container.ID[:12]), 3*time.Second)
							helper.UpdateToggleText(currentFilteredState, statusToggle)
							refreshTable(currentFilteredState)
						}
						return nil
					case 'd': // Remove container
						forceRemove := strings.HasPrefix(container.Status, "Up")
						if forceRemove {
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

					case 'i': // Inspect container - opens the modal
						// displayTimedStatus(fmt.Sprintf("Inspecting container %s...", container.ID[:12]), 2*time.Second)
						go func() {
							defer func() {
								if r := recover(); r != nil {
									// fmt.Fprintf(os.Stderr, "DEBUG: Recovered from panic in inspect goroutine: %v\n", r)
									tuiApp.QueueUpdateDraw(func() {
										displayTimedStatus(fmt.Sprintf("Internal error: %v", r), 5*time.Second)
										if currentPageName, _ := pages.GetFrontPage(); currentPageName == "inspect_view" {
											pages.RemovePage("inspect_view")
										}
										tuiApp.SetFocus(table)
									})
								}
							}()

							cID := container.ID
							ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Increased timeout
							defer cancel()

							inspectRaw, err := usecase.Query.ContainerInspect(ctx, cID)
							if err != nil {
								fmt.Fprintf(os.Stderr, "DEBUG: Failed to inspect container: %v\n", err)
								tuiApp.QueueUpdateDraw(func() { // IMPORTANT: Ensure UI update for error
									displayTimedStatus(fmt.Sprintf("Inspect error: %v", err), 7*time.Second) // Longer display for error
									tuiApp.SetFocus(table)                                                   // Ensure focus returns to main table
								})
								return // Stop processing if Docker inspect fails
							}

							var pretty bytes.Buffer
							err = json.Indent(&pretty, []byte(inspectRaw), "", "  ")
							if err != nil {
								// fmt.Fprintf(os.Stderr, "DEBUG: Failed to format JSON: %v\n", err)
								pretty.WriteString(inspectRaw) // fallback to raw
								tuiApp.QueueUpdateDraw(func() {
									displayTimedStatus("JSON format error; showing raw data.", 3*time.Second)
								})
							}
							textView := tview.NewTextView().
								SetScrollable(true).
								SetWordWrap(false).
								SetText(pretty.String())

							// SetBackgroundColor returns *tview.Box, so chain breaks here.
							textView.SetBackgroundColor(tcell.ColorBlack)
							// SetTextColor must be called directly on textView.
							textView.SetTextColor(tcell.ColorWhite)

							textView.SetBorder(true).
								SetBorderColor(tcell.ColorGreen)

							// Assuming container.Names[0] needs to be converted to string based on prior feedback.
							textView.SetTitle(fmt.Sprintf(" Inspecting %s ", string(container.Name))).
								SetTitleAlign(tview.AlignCenter)

							// Removed the problematic tview.NewModal().SetContent(textView) part.
							closeButton := tview.NewButton("Close [Ctrl + C]").SetSelectedFunc(func() {
								tuiApp.QueueUpdateDraw(func() {
									pages.RemovePage("inspect_view")
									tuiApp.SetFocus(table) // Return focus to the table
								})
							})

							inspectContentFlex := tview.NewFlex().SetDirection(tview.FlexRow).
								AddItem(textView, 0, 1, true).    // TextView takes most space
								AddItem(closeButton, 1, 0, false) // Button at the bottom

							// This is the main Flex that acts as your modal's frame and centers it.
							centeredModal := tview.NewFlex().
								AddItem(nil, 0, 1, false). // Left padding
								AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
									AddItem(nil, 0, 1, false).               // Top padding
									AddItem(inspectContentFlex, 0, 3, true). // The actual content, taking 3/5 of vertical space
									AddItem(nil, 0, 1, false),               // Bottom padding
												0, 3, false). // The content flex, taking 3/5 of horizontal space
								AddItem(nil, 0, 1, false) // Right padding

							tuiApp.QueueUpdateDraw(func() {
								// Add and show this Flex-based modal as a new page
								pages.AddPage("inspect_view", centeredModal, true, true)
								// fmt.Fprintf(os.Stderr, "DEBUG: Setting focus to TextView after AddPage\n")
								tuiApp.SetFocus(textView)
							})
						}()
						return nil // Consume 'i' event for the main table
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
		}
		return event
	})

	tuiApp.SetRoot(pages, true).EnableMouse(true).SetFocus(table)

	if err := tuiApp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}

	defer stopAutoRefreshFunc()
}

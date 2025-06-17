package handler

import (
	"bytes"
	"context"

	"encoding/json"
	"strings"
	"time"

	"fmt"
	"os"

	"docker-tui/internal/app"
	"docker-tui/internal/domain"
	"docker-tui/internal/helper"
	"docker-tui/internal/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ContainerUseCases = app.ContainerUseCases

// Global variables for tview components (accessible across RunCLI and its closures)
var tuiApp *tview.Application
var pages *tview.Pages
var table *tview.Table

// var mainAppLayout *tview.Flex // Your main app layout containing table, status bar, etc.

var autoRefreshInterval = 3 * time.Second
var autoRefreshEnabled = false
var stopAutoRefresh chan struct{}

const (
	PageMainMenu      = "main_menu"
	PageContainerList = "container_list"
	PageInspectView   = "inspect_view" // Constant for the inspect modal page name
)

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

	// Initialize pages
	pages = tview.NewPages()

	// Function to refresh the container table
	refreshTable := func(state string) {
		currentFilteredState = state
		containers, err := usecase.Query.ListContainersByState(context.Background(), state)
		if err != nil {
			helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Error fetching containers: %v", err), 3*time.Second, statusBar, defaultBarText)
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
		helper.DisplayTimedStatus(
			tuiApp,
			fmt.Sprintf("Auto-refresh enabled (every %s). Press 'a' to disable.",
				autoRefreshInterval),
			3*time.Second, statusBar, defaultBarText)
	}

	stopAutoRefreshFunc := func() {
		if !autoRefreshEnabled {
			return
		}
		autoRefreshEnabled = false
		close(stopAutoRefresh)
		stopAutoRefresh = make(chan struct{})
		helper.DisplayTimedStatus(
			tuiApp,
			"Auto-refresh disabled. Press 'a' to enable.",
			3*time.Second,
			statusBar,
			defaultBarText,
		)
	}

	mainMenuPage := ui.CreateMainMenu(tuiApp, pages, table)
	containerListPage := ui.CreateContainerListPage(tuiApp, table, statusToggle, statusBar, exitGuide)

	pages.AddPage(PageMainMenu, mainMenuPage, true, true)
	// Add the container list page, initially hidden
	pages.AddPage(PageContainerList, containerListPage, true, false)

	// Initial refresh and start auto-refresh
	refreshTable(currentFilteredState)
	startAutoRefresh()

	tuiApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentPageName, _ := pages.GetFrontPage()

		if currentPageName == "inspect_view" {
			return event
		}

		// Main application input capture logic (only active when "main" page is front)
		switch currentPageName {
		case PageMainMenu:
			// fmt.Fprintf(os.Stderr, "DEBUG: Handling input for main menu.\n")
			// For button-based menu, arrow keys/Tab/Enter are handled by tview.Flex and Buttons.
			// Explicit 'Escape' to quit if 'q' is only handled by the button.
			if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
				tuiApp.Stop()
				return nil // Consume event
			}
			return event // Propagate to mainMenuPage for its internal handling

		case PageContainerList:
			// fmt.Fprintf(os.Stderr, "DEBUG: Handling input for container list.\n")
			// Handle keys to go back to the Main Menu
			if event.Rune() == 'm' { // 'm' for Menu
				tuiApp.QueueUpdateDraw(func() {
					pages.SwitchToPage(PageMainMenu)
					tuiApp.SetFocus(mainMenuPage) // Set focus back to the menu list
				})
				return nil // Consume event
			}
			// Add 'Escape' as an additional way to go back from container list to main menu
			if event.Key() == tcell.KeyEscape {
				tuiApp.QueueUpdateDraw(func() {
					pages.SwitchToPage(PageMainMenu)
					tuiApp.SetFocus(mainMenuPage)
				})
				return nil
			}

			switch event.Key() { // This switch handles tcell.Key constants (Tab, Enter)
			case tcell.KeyTab:
				if tuiApp.GetFocus() == statusToggle {
					tuiApp.SetFocus(table)
					helper.DisplayTimedStatus(tuiApp, "Focus: table", 1*time.Second, statusBar, defaultBarText)
				} else {
					tuiApp.SetFocus(statusToggle)
					helper.DisplayTimedStatus(tuiApp, "Focus: statusToggle", 1*time.Second, statusBar, defaultBarText)
				}
				return nil
			case tcell.KeyEnter:
				if tuiApp.GetFocus() == statusToggle {
					if currentFilteredState == "running" {
						refreshTable("exited")
					} else {
						refreshTable("running")
					}
					helper.DisplayTimedStatus(
						tuiApp,
						fmt.Sprintf("Switched to: %s", currentFilteredState),
						2*time.Second, statusBar, defaultBarText)
					return nil
				}
			default:
				if event.Rune() == 'q' { // 'q' to quit (main app)
					tuiApp.Stop()
					return nil
				}

				if tuiApp.GetFocus() == table {
					row, _ := table.GetSelection()
					if row < 1 { // Header row check
						helper.DisplayTimedStatus(tuiApp, "No container selected (header row).", 2*time.Second, statusBar, defaultBarText)
						return nil
					}
					cell := table.GetCell(row, 0)
					if cell == nil || cell.GetReference() == nil {
						helper.DisplayTimedStatus(tuiApp, "Error: No container reference found.", 3*time.Second, statusBar, defaultBarText)
						return nil
					}
					container := cell.GetReference().(*domain.Container)

					switch event.Rune() {
					case 's': // Start container
						err := usecase.Control.StartContainer(context.Background(), container.ID)
						if err != nil {
							helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Failed to start: %v", err), 5*time.Second, statusBar, defaultBarText)
						} else {
							helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Container %s started. Switched to Active view.", container.ID[:12]), 3*time.Second, statusBar, defaultBarText)
							currentFilteredState = "running"
							helper.UpdateToggleText(currentFilteredState, statusToggle)
							refreshTable(currentFilteredState)
						}
						return nil
					case 'x': // Stop container
						err := usecase.Control.StopContainer(context.Background(), container.ID)
						if err != nil {
							helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Failed to stop: %v", err), 5*time.Second, statusBar, defaultBarText)
						} else {
							helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Container %s stopped.", container.ID[:12]), 3*time.Second, statusBar, defaultBarText)
							helper.UpdateToggleText(currentFilteredState, statusToggle)
							refreshTable(currentFilteredState)
						}
						return nil
					case 'r': // Restart container
						err := usecase.Control.RestartContainer(context.Background(), container.ID)
						if err != nil {
							helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Failed to restart: %v", err), 5*time.Second, statusBar, defaultBarText)
						} else {
							helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Container %s restarted.", container.ID[:12]), 3*time.Second, statusBar, defaultBarText)
							helper.UpdateToggleText(currentFilteredState, statusToggle)
							refreshTable(currentFilteredState)
						}
						return nil
					case 'd': // Remove container
						forceRemove := strings.HasPrefix(container.Status, "Up")
						if forceRemove {
							helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Attempting to force remove running container %s...", container.ID[:12]), 2*time.Second, statusBar, defaultBarText)
						} else {
							helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Attempting to remove stopped container %s...", container.ID[:12]), 2*time.Second, statusBar, defaultBarText)
						}
						err := usecase.Control.RemoveContainer(context.Background(), container.ID, forceRemove)
						if err != nil {
							helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Failed to remove: %v", err), 5*time.Second, statusBar, defaultBarText)
						} else {
							helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Container %s removed.", container.ID[:12]), 3*time.Second, statusBar, defaultBarText)
							helper.UpdateToggleText(currentFilteredState, statusToggle)
							refreshTable(currentFilteredState)
						}
						return nil
					case 'i': // Inspect container - opens the modal
						go func() {
							defer func() {
								if r := recover(); r != nil {
									tuiApp.QueueUpdateDraw(func() {
										helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Internal error: %v", r), 5*time.Second, statusBar, defaultBarText)
										if currentPageName, _ := pages.GetFrontPage(); currentPageName == "inspect_view" {
											pages.RemovePage("inspect_view")
										}
										tuiApp.SetFocus(table)
									})
								}
							}()

							cID := container.ID
							ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
							defer cancel()

							inspectRaw, err := usecase.Query.ContainerInspect(ctx, cID)
							if err != nil {
								fmt.Fprintf(os.Stderr, "DEBUG: Failed to inspect container: %v\n", err)
								tuiApp.QueueUpdateDraw(func() {
									helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Inspect error: %v", err), 7*time.Second, statusBar, defaultBarText)
									tuiApp.SetFocus(table)
								})
								return
							}

							var pretty bytes.Buffer
							err = json.Indent(&pretty, []byte(inspectRaw), "", "  ")
							if err != nil {
								pretty.WriteString(inspectRaw)
								tuiApp.QueueUpdateDraw(func() {
									helper.DisplayTimedStatus(tuiApp, "JSON format error; showing raw data.", 3*time.Second, statusBar, defaultBarText)
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
							closeAction := func() {
								pages.RemovePage("inspect_view")
								tuiApp.SetFocus(table)
							}
							closeButton := tview.NewButton("Close").SetSelectedFunc(closeAction)

							buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn).
								AddItem(closeButton, 15, 0, true). // Adjust '15' for desired width.
								AddItem(nil, 0, 1, false)

							inspectContentFlex := tview.NewFlex().SetDirection(tview.FlexRow).
								AddItem(textView, 0, 1, true).        // TextView takes most space
								AddItem(buttonContainer, 1, 0, false) // Button at the bottom

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

				if currentPageName == PageContainerList {
					if event.Rune() == 'm' { // 'm' for Menu
						pages.SwitchToPage(PageMainMenu)
						tuiApp.SetFocus(mainMenuPage)
						return nil
					}
					if event.Key() == tcell.KeyEscape {
						pages.SwitchToPage(PageMainMenu)
						tuiApp.SetFocus(mainMenuPage)
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

package handler

import (
	"time"

	"fmt"
	"os"

	"docker-tui/internal/app"
	"docker-tui/internal/helper"
	"docker-tui/internal/ui/containers"
	"docker-tui/internal/ui/menu"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type AppUseCases struct {
	Containers *app.ContainerUseCases
}

// Global variables for tview components (accessible across RunCLI and its closures)
var tuiApp *tview.Application
var pages *tview.Pages
var containerTable *tview.Table

// var mainAppLayout *tview.Flex // Your main app layout containing table, status bar, etc.

var autoRefreshInterval = 3 * time.Second
var autoRefreshEnabled = false
var stopAutoRefresh chan struct{}

const (
	PageMainMenu      = "main_menu"
	PageContainerList = "container_list"
	PageInspectView   = "inspect_view" // Constant for the inspect modal page name
)

func RunCLI(usecases *AppUseCases) {
	tuiApp = tview.NewApplication()

	var currentContainerFilterState string = "running"
	const defaultBarText = ""

	containerTable = helper.TableFormat() // Initialize the table

	statusToggle := helper.StatusToggle(tuiApp)
	statusBar := helper.StatusBar()
	exitGuide := helper.ExitGuide()

	// Initialize UI elements
	// helper.UpdateToggleText(currentContainerFilterState, statusToggle)

	stopAutoRefresh = make(chan struct{})

	// Initialize pages
	pages = tview.NewPages()

	displayTimedStatus := func(message string, duration time.Duration) {
		helper.DisplayTimedStatus(tuiApp, message, duration, statusBar, "") // defaultBarText is ""
	}

	var refreshContainerTable func(state string)

	// Function to refresh the container table
	// refreshContainerTable := func(state string) {
	// 	currentContainerFilterState = state
	// 	containers, err := usecases.Containers.Query.ListContainersByState(context.Background(), state)
	// 	if err != nil {
	// 		helper.DisplayTimedStatus(tuiApp, fmt.Sprintf("Error fetching containers: %v", err), 3*time.Second, statusBar, defaultBarText)
	// 		return
	// 	}
	// 	helper.PopulateContainerTableUI(containerTable, containers)
	// 	helper.UpdateToggleText(state, statusToggle)
	// }

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
					currentPageName, _ := pages.GetFrontPage()
					if currentPageName == PageContainerList {
						// Check if refreshContainerTable is initialized (safety)
						if refreshContainerTable != nil {
							refreshContainerTable(currentContainerFilterState)
						}
					}
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
		helper.DisplayTimedStatus(
			tuiApp,
			"Auto-refresh disabled. Press 'a' to enable.",
			3*time.Second,
			statusBar,
			defaultBarText,
		)
	}

	closeInspectModal := func() {
		fmt.Fprintf(os.Stderr, "DEBUG: Attempting to close inspect_view.\n")
		fmt.Fprintf(os.Stderr, "DEBUG: Inside QueueUpdateDraw for closing inspect_view (before remove).\n")
		pages.RemovePage(PageInspectView)
		tuiApp.SetFocus(containerTable) // Always return focus to the container table after inspect
		fmt.Fprintf(os.Stderr, "DEBUG: inspect_view removed, focus set to table.\n")
	}
	mainMenuPage := menu.CreateMainMenu(tuiApp, pages, containerTable, displayTimedStatus)

	containerPageConfig := containers.Config{
		App:                   tuiApp,
		Pages:                 pages,
		Table:                 containerTable,
		StatusToggle:          statusToggle,
		StatusBar:             statusBar,
		ExitGuide:             exitGuide, // Potentially change to "Back to Menu"
		UseCases:              usecases.Containers,
		DisplayStatus:         displayTimedStatus,
		CurrentFilterState:    &currentContainerFilterState, // Pass pointer to allow modification
		CloseInspectModalFunc: closeInspectModal,            // Pass the modal close function
		StartAutoRefreshFunc:  startAutoRefresh,             // Pass auto-refresh controls
		StopAutoRefreshFunc:   stopAutoRefreshFunc,
	}

	containerListPage := containers.NewContainerListPage(containerPageConfig) // NewContainerListPage now defines its own refreshTableFunc

	pages.AddPage(PageMainMenu, mainMenuPage, true, true)
	pages.AddPage(PageContainerList, containerListPage, true, false)

	refreshContainerTable = containerListPage.RefreshTable
	// Initial refresh and start auto-refresh
	refreshContainerTable(currentContainerFilterState)
	startAutoRefresh()

	tuiApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentPageName, _ := pages.GetFrontPage()

		switch currentPageName {
		case PageMainMenu:
			if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
				tuiApp.Stop()
				return nil
			}
			return event // Propagate to mainMenuPage (Flex/Buttons)

		case PageContainerList:
			// Delegate input handling to the ContainerListPage's specific method
			return containerListPage.HandleInput(event)

		case PageInspectView: // Handle keys specifically for the inspect modal
			switch event.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				closeInspectModal()
				return nil
			default:
				if event.Rune() == 'q' || event.Rune() == 'c' {
					closeInspectModal()
					return nil
				}
			}
			return event // Propagate to TextView for scrolling

		default: // Global fallback for keys not handled by any specific page
			if event.Key() == tcell.KeyCtrlC || event.Rune() == 'q' {
				tuiApp.Stop()
				return nil
			}
		}

		return event // Ensure unhandled events propagate if not explicitly handled
	})

	tuiApp.SetRoot(pages, true).EnableMouse(true).SetFocus(mainMenuPage)

	if err := tuiApp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}

	defer stopAutoRefreshFunc()
}

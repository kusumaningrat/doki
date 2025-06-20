package handler

import (
	"time"

	"fmt"
	"os"

	"docker-tui/internal/app"
	"docker-tui/internal/helper"
	"docker-tui/internal/ui/containers"
	"docker-tui/internal/ui/images"
	"docker-tui/internal/ui/menu"
	"docker-tui/internal/ui/networks"
	"docker-tui/internal/ui/volumes"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type AppUseCases struct {
	Containers *app.ContainerUseCases
	Images     *app.ImageUseCases
	Volumes    *app.VolumeUseCases
	Networks   *app.NetworkUseCases
}

// Global variables for tview components (accessible across RunCLI and its closures)
var tuiApp *tview.Application
var pages *tview.Pages
var containerTable *tview.Table
var imageTable *tview.Table
var volumeTable *tview.Table
var networkTable *tview.Table

// var mainAppLayout *tview.Flex // Your main app layout containing table, status bar, etc.

var autoRefreshInterval = 3 * time.Second
var autoRefreshEnabled = false
var stopAutoRefresh chan struct{}

const (
	PageMainMenu    = "main_menu"
	PageInspectView = "inspect_view"

	PageContainerList = "container_list"
	PageImageList     = "image_list"
	PageVolumeList    = "volume_list"
	PageNetworkList   = "network_list"
)

func RunCLI(usecases *AppUseCases) {
	tuiApp = tview.NewApplication()

	var currentContainerFilterState string = "running"
	const defaultBarText = ""

	containerTable = helper.ContainerTableFormat()
	imageTable = helper.ImageTableFormat()
	volumeTable = helper.VolumeTableFormat()
	networkTable = helper.NetworkTableFormat()

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
	var refreshImageTable func()
	var refreshVolumeTable func()
	var refreshNetworkTable func()

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
		pages.RemovePage(PageInspectView)
		tuiApp.SetFocus(containerTable) // Always return focus to the container table after inspect
	}
	mainMenuPage := menu.CreateMainMenu(
		tuiApp,
		pages,
		containerTable,
		imageTable,
		volumeTable,
		networkTable,
		displayTimedStatus,
	)

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

	imageConfig := images.Config{
		App:                  tuiApp,
		Pages:                pages,
		Table:                imageTable,
		StatusBar:            statusBar,
		ExitGuide:            exitGuide, // Potentially change to "Back to Menu"
		DisplayStatus:        displayTimedStatus,
		StartAutoRefreshFunc: startAutoRefresh, // Pass auto-refresh controls
		StopAutoRefreshFunc:  stopAutoRefreshFunc,
		UseCases:             usecases.Images,
	}

	volumeConfig := volumes.Config{
		App:                  tuiApp,
		Pages:                pages,
		Table:                volumeTable,
		StatusBar:            statusBar,
		ExitGuide:            exitGuide,
		DisplayStatus:        displayTimedStatus,
		StartAutoRefreshFunc: startAutoRefresh,
		StopAutoRefreshFunc:  stopAutoRefreshFunc,
		UseCases:             usecases.Volumes,
	}

	networkConfig := networks.Config{
		App:                  tuiApp,
		Pages:                pages,
		Table:                networkTable,
		StatusBar:            statusBar,
		ExitGuide:            exitGuide,
		DisplayStatus:        displayTimedStatus,
		StartAutoRefreshFunc: startAutoRefresh,
		StopAutoRefreshFunc:  stopAutoRefreshFunc,
		UseCases:             usecases.Networks,
	}

	containerListPage := containers.NewContainerListPage(containerPageConfig)
	imageListPage := images.NewImageListPage(imageConfig)
	volumeListPage := volumes.NewVolumeListPage(volumeConfig)
	networkListPage := networks.NewNetworkListPage(networkConfig)

	pages.AddPage(PageMainMenu, mainMenuPage, true, true)
	pages.AddPage(PageImageList, imageListPage, true, false)
	pages.AddPage(PageVolumeList, volumeListPage, true, false)
	pages.AddPage(PageNetworkList, networkListPage, true, false)

	refreshContainerTable = containerListPage.RefreshTable
	refreshImageTable = imageListPage.RefreshTable
	refreshVolumeTable = volumeListPage.RefreshTable
	refreshNetworkTable = networkListPage.RefreshTable
	// Initial refresh and start auto-refresh
	refreshContainerTable(currentContainerFilterState)
	refreshImageTable()
	refreshVolumeTable()
	refreshNetworkTable()
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

		case PageImageList:
			// Delegate input handling to the ContainerListPage's specific method
			return imageListPage.HandleInput(event)

		case PageVolumeList:
			// Delegate input handling to the ContainerListPage's specific method
			return volumeListPage.HandleInput(event)

		case PageNetworkList:
			// Delegate input handling to the ContainerListPage's specific method
			return networkListPage.HandleInput(event)

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

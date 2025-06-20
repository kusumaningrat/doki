package containers

import (
	"context"
	"docker-tui/internal/app"
	"docker-tui/internal/domain"
	"docker-tui/internal/helper"
	"docker-tui/internal/ui/inspect"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	PageMainMenu      = "main_menu"
	PageContainerList = "container_list"
	PageInspectView   = "inspect_view"
	// Other pages like PageImageList, etc.
)

type Config struct {
	App                   *tview.Application
	Pages                 *tview.Pages
	Table                 *tview.Table // The actual container table
	StatusToggle          *tview.TextView
	StatusBar             *tview.TextView
	ExitGuide             *tview.TextView // This might change to a "Back to menu" guide
	UseCases              *app.ContainerUseCases
	DisplayStatus         func(message string, duration time.Duration)
	CurrentFilterState    *string
	CloseInspectModalFunc func()
	StartAutoRefreshFunc  func()
	StopAutoRefreshFunc   func()
	AutoRefreshEnabled    bool
	SetFocusOnCloseModal  func(tview.Primitive)
}

type ContainerListPage struct {
	*tview.Flex // Embeds the tview.Flex for its layout
	config      Config
}

func (p *ContainerListPage) RefreshTable(state string) {
	*p.config.CurrentFilterState = state // Update filter state
	containers, err := p.config.UseCases.Query.ListContainersByState(context.Background(), state)
	if err != nil {
		p.config.DisplayStatus(fmt.Sprintf("Error fetching containers: %v", err), 3*time.Second)
		return
	}
	helper.PopulateContainerTableUI(p.config.Table, containers)
	helper.UpdateToggleText(*p.config.CurrentFilterState, p.config.StatusToggle)
}

// NewContainerListPage creates and returns the containers page.
func NewContainerListPage(cfg Config) *ContainerListPage {
	// Assign the passed refresh function to the struct
	p := &ContainerListPage{
		Flex:   tview.NewFlex().SetDirection(tview.FlexRow), // Initialize Flex below
		config: cfg,
	}

	// Initial refresh when page is created by calling the new public method
	p.RefreshTable(*p.config.CurrentFilterState) // Call the public method

	// Define the layout (Flex) and assign it to p.Flex
	p.Flex.
		AddItem(p.config.StatusToggle, 1, 0, false).
		AddItem(p.config.Table, 0, 1, true). // Table is the primary focus of this view
		AddItem(p.config.StatusBar, 1, 0, false).
		AddItem(p.config.ExitGuide, 2, 0, false) // This guide might instruct to press 'm' for menu

	return p
}

// HandleInput is the page-specific input handler for the ContainerListPage.
// It takes the event and other necessary global primitives/functions as parameters.
func (p *ContainerListPage) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	// Handle keys to go back to the Main Menu
	if event.Rune() == 'm' { // 'm' for Menu
		p.config.Pages.SwitchToPage(PageMainMenu)
		_, primitive := p.config.Pages.GetFrontPage()
		p.config.App.SetFocus(primitive)
		return nil // Consume event
	}
	if event.Key() == tcell.KeyEscape {
		p.config.Pages.SwitchToPage(PageMainMenu)
		_, primitive := p.config.Pages.GetFrontPage()
		p.config.App.SetFocus(primitive)
		return nil
	}

	// Container list specific keybindings (s, x, r, d, i, a, Tab, Enter)
	switch event.Key() { // Handles tcell.Key constants (Tab, Enter, Escape, CtrlC)
	case tcell.KeyTab:
		if p.config.App.GetFocus() == p.config.StatusToggle {
			p.config.App.SetFocus(p.config.Table)
			p.config.DisplayStatus("Focus: table", 1*time.Second)
		} else {
			p.config.App.SetFocus(p.config.StatusToggle)
			p.config.DisplayStatus("Focus: statusToggle", 1*time.Second)
		}
		return nil
	case tcell.KeyEnter:
		if p.config.App.GetFocus() == p.config.StatusToggle {
			nextState := "running" // Default next state
			if *p.config.CurrentFilterState == "running" {
				nextState = "exited" // If currently running, switch to exited
			}
			// Update the current filter state variable
			*p.config.CurrentFilterState = nextState // Ensure this is updated before refreshing
			p.RefreshTable(nextState)                // Call the method directly on 'p'
			p.config.DisplayStatus(fmt.Sprintf("Switched to: %s", *p.config.CurrentFilterState), 2*time.Second)
			return nil
		}
	default: // Handles rune-based keys (s, x, r, d, i, a, q)
		if event.Rune() == 'q' { // 'q' to quit (global app exit from this page)
			p.config.App.Stop()
			return nil
		}

		// Actions that require a container to be selected (s, x, r, d, i, a)
		if p.config.App.GetFocus() == p.config.Table { // Only proceed if table is focused
			row, _ := p.config.Table.GetSelection()
			if row < 1 { // Header row check
				p.config.DisplayStatus("No container selected (header row).", 2*time.Second)
				return nil
			}
			cell := p.config.Table.GetCell(row, 0)
			if cell == nil || cell.GetReference() == nil {
				p.config.DisplayStatus("Error: No container reference found.", 3*time.Second)
				return nil
			}
			container := cell.GetReference().(*domain.Container)

			switch event.Rune() {
			case 's': // Start container
				p.config.DisplayStatus(fmt.Sprintf("Starting container %s...", container.ID[:12]), 5*time.Second)
				go func() { // Perform Docker action in a goroutine
					err := p.config.UseCases.Control.StartContainer(context.Background(), container.ID)
					p.config.App.QueueUpdateDraw(func() { // Queue UI update
						if err != nil {
							p.config.DisplayStatus(fmt.Sprintf("Failed to start: %v", err), 5*time.Second)
						} else {
							p.config.DisplayStatus(fmt.Sprintf("Container %s started. Switched to Active view.", container.ID[:12]), 5*time.Second)
							*p.config.CurrentFilterState = "running" // Update state
							helper.UpdateToggleText(*p.config.CurrentFilterState, p.config.StatusToggle)
							// p.refreshTableFunc(*p.config.CurrentFilterState) // Refresh table
							p.config.App.SetFocus(p.config.Table)
						}
					})
				}()
				return nil
			case 'x': // Stop container
				p.config.DisplayStatus(fmt.Sprintf("Stopping container %s...", container.ID[:12]), 5*time.Second)
				go func() { // Perform Docker action in a goroutine
					err := p.config.UseCases.Control.StopContainer(context.Background(), container.ID)
					p.config.App.QueueUpdateDraw(func() { // Queue UI update
						if err != nil {
							p.config.DisplayStatus(fmt.Sprintf("Failed to stop: %v", err), 5*time.Second)
						} else {
							p.config.DisplayStatus(fmt.Sprintf("Container %s stopped.", container.ID[:12]), 3*time.Second)
							// p.refreshTableFunc(*p.config.CurrentFilterState) // Refresh table
							p.config.App.SetFocus(p.config.Table)
						}
					})
				}()
				return nil
			case 'r': // Restart container
				p.config.DisplayStatus(fmt.Sprintf("Restarting container %s...", container.ID[:12]), 5*time.Second)
				go func() { // Perform Docker action in a goroutine
					err := p.config.UseCases.Control.RestartContainer(context.Background(), container.ID)
					p.config.App.QueueUpdateDraw(func() { // Queue UI update
						if err != nil {
							p.config.DisplayStatus(fmt.Sprintf("Failed to restart: %v", err), 5*time.Second)
						} else {
							p.config.DisplayStatus(fmt.Sprintf("Container %s restarted.", container.ID[:12]), 5*time.Second)
							// p.refreshTableFunc(*p.config.CurrentFilterState) // Refresh table
							p.config.App.SetFocus(p.config.Table)
						}
					})
				}()
				return nil
			case 'd': // Remove container
				p.config.DisplayStatus(fmt.Sprintf("Attempting to remove container %s...", container.ID[:12]), 5*time.Second)
				go func() { // Perform Docker action in a goroutine
					forceRemove := strings.HasPrefix(container.Status, "Up")
					err := p.config.UseCases.Control.RemoveContainer(context.Background(), container.ID, forceRemove)
					p.config.App.QueueUpdateDraw(func() { // Queue UI update
						if err != nil {
							p.config.DisplayStatus(fmt.Sprintf("Failed to remove: %v", err), 5*time.Second)
						} else {
							p.config.DisplayStatus(fmt.Sprintf("Container %s removed.", container.ID[:12]), 5*time.Second)
							// p.refreshTableFunc(*p.config.CurrentFilterState) // Refresh table
							p.config.App.SetFocus(p.config.Table)
						}
					})
				}()
				return nil
			case 'i': // Inspect container - opens the modal
				p.config.DisplayStatus(fmt.Sprintf("Inspecting container %s...", container.ID[:12]), 5*time.Second)
				go func(selectedContainer *domain.Container) { // Pass container by value
					defer func() {
						if r := recover(); r != nil {
							p.config.App.QueueUpdateDraw(func() {
								p.config.DisplayStatus(fmt.Sprintf("Internal error: %v", r), 5*time.Second)
								if currentPageName, _ := p.config.Pages.GetFrontPage(); currentPageName == PageInspectView {
									p.config.Pages.RemovePage(PageInspectView)
								}
								p.config.App.SetFocus(p.config.Table)
							})
						}
					}()

					cID := selectedContainer.ID
					ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()

					inspectRaw, err := p.config.UseCases.Query.ContainerInspect(ctx, cID)
					if err != nil {
						p.config.App.QueueUpdateDraw(func() {
							p.config.DisplayStatus(fmt.Sprintf("Inspect error: %v", err), 5*time.Second)
							p.config.App.SetFocus(p.config.Table)
						})
						return
					}

					// Create the inspect modal content and add it as a new page
					p.config.SetFocusOnCloseModal(p.config.Table)
					inspectModalContent := inspect.CreateInspectModal(p.config.App, p.config.Pages, p.config.Table, inspectRaw, selectedContainer.Name, p.config.DisplayStatus)

					p.config.App.QueueUpdateDraw(func() {
						p.config.Pages.AddPage(PageInspectView, inspectModalContent, true, true)
						p.config.App.SetFocus(inspectModalContent)
					})
				}(container) // Pass the selected container to the goroutine
				return nil
			case 'a': // 'a' for auto-refresh toggle
				if p.config.AutoRefreshEnabled { // Check autoRefreshEnabled from config
					p.config.StopAutoRefreshFunc()
				} else {
					p.config.StartAutoRefreshFunc()
				}
				return nil
			}
		}
		return event
	}
	return event
}

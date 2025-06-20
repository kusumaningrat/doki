package volumes

import (
	// For Docker API calls
	"context"
	"fmt"
	"time" // For time.Duration

	"docker-tui/internal/app"    // Your application use cases
	"docker-tui/internal/domain" // Your domain models
	"docker-tui/internal/helper"
	"docker-tui/internal/ui/inspect"

	// Your general UI helpers
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	PageMainMenu    = "main_menu"
	PageVolumeList  = "volume_list"
	PageInspectView = "inspect_view"
)

type Config struct {
	App                  *tview.Application
	Pages                *tview.Pages
	Table                *tview.Table
	StatusBar            *tview.TextView
	ExitGuide            *tview.TextView
	UseCases             *app.VolumeUseCases
	DisplayStatus        func(message string, duration time.Duration)
	StartAutoRefreshFunc func()
	StopAutoRefreshFunc  func()
	SetFocusOnCloseModal func(tview.Primitive)
}

type VolumeListPage struct {
	*tview.Flex      // Embeds the tview.Flex for its layout
	config           Config
	refreshTableFunc func() // Function to refresh this page's table
}

func (p *VolumeListPage) RefreshTable() {
	volumes, err := p.config.UseCases.Query.ListAllVolumes(context.Background())

	if err != nil {
		p.config.DisplayStatus(fmt.Sprintf("Error fetching volumes: %v", err), 3*time.Second)
		return
	}

	// Now, wrap the UI update part in QueueUpdateDraw
	helper.PopulateVolumeTableUI(p.config.Table, volumes)
}

func NewVolumeListPage(cfg Config) *VolumeListPage {
	// Initial refresh when page is created (optional, or call on switch)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Docker Volumes").SetTextAlign(tview.AlignCenter), 1, 0, false).
		AddItem(cfg.Table, 0, 1, true).
		AddItem(cfg.StatusBar, 1, 0, false).
		AddItem(cfg.ExitGuide, 1, 0, false)
	page := &VolumeListPage{ // Create the VolumeListPage instance
		Flex:   layout,
		config: cfg,
	}

	page.refreshTableFunc = func() {
		page.RefreshTable()
	}

	return page
}

// HandleInput is the page-specific input handler for the VolumeListPage.
func (p *VolumeListPage) HandleInput(
	event *tcell.EventKey) *tcell.EventKey { // Changed params

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

	switch event.Key() { // Handles tcell.Key constants
	// No Tab/Enter for toggle like containers, unless you add a filter
	default: // Handles rune-based keys
		if event.Rune() == 'q' { // 'q' to quit (global app exit from this page)
			p.config.App.Stop()
			return nil
		}

		if p.config.App.GetFocus() == p.config.Table { // Only proceed if table is focused
			row, _ := p.config.Table.GetSelection()
			if row < 1 { // Header row check
				p.config.DisplayStatus("No volume selected (header row).", 2*time.Second)
				return nil
			}
			cell := p.config.Table.GetCell(row, 0)
			if cell == nil || cell.GetReference() == nil {
				p.config.DisplayStatus("Error: No volume reference found.", 3*time.Second)
				return nil
			}
			volume := cell.GetReference().(*domain.Volume)

			switch event.Rune() {
			case 'r':
				p.config.DisplayStatus(fmt.Sprintf("Removing volume %s:%s...", volume.Name, volume.Driver), 1*time.Second)

				go func() { // Perform Docker action in a goroutine
					err := p.config.UseCases.Control.RemoveVolume(context.Background(), volume.Name)

					p.config.App.QueueUpdateDraw(func() {
						if err != nil {
							p.config.DisplayStatus(fmt.Sprintf("Failed to remove: %v", err), 10*time.Second)
						} else {
							p.config.DisplayStatus(fmt.Sprintf("Volume %s removed.", volume.Name), 3*time.Second)
						}
						p.refreshTableFunc()
						p.config.App.SetFocus(p.config.Table)

					})

				}()
				return nil

			case 'i':
				p.config.DisplayStatus(fmt.Sprintf("Inspecting volume %s...", volume.Name[:12]), 1*time.Second)
				go func(selectedVolume *domain.Volume) {
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

					volumeName := selectedVolume.Name
					ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()

					inspectRaw, err := p.config.UseCases.Query.VolumeInspect(ctx, volumeName)
					if err != nil {
						p.config.App.QueueUpdateDraw(func() {
							p.config.DisplayStatus(fmt.Sprintf("Inspect error: %v", err), 7*time.Second)
							p.config.App.SetFocus(p.config.Table)
						})
						return
					}

					p.config.SetFocusOnCloseModal(p.config.Table)
					// Create the inspect modal content and add it as a new page
					inspectModalContent := inspect.CreateInspectModal(p.config.App, p.config.Pages, p.config.Table, inspectRaw, selectedVolume.Name, p.config.DisplayStatus)

					p.config.App.QueueUpdateDraw(func() {
						p.config.Pages.AddPage(PageInspectView, inspectModalContent, true, true)
						p.config.App.SetFocus(inspectModalContent.GetContentPrimitive())
					})
				}(volume) // Pass the selected container to the goroutine
				return nil
			}

		}
	}
	return event
}

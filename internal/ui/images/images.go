package images

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
	PageImageList   = "image_list" // NEW
	PageInspectView = "inspect_view"
)

type Config struct {
	App                  *tview.Application
	Pages                *tview.Pages
	Table                *tview.Table
	StatusBar            *tview.TextView
	ExitGuide            *tview.TextView
	UseCases             *app.ImageUseCases
	DisplayStatus        func(message string, duration time.Duration)
	StartAutoRefreshFunc func() // If image list also supports auto-refresh
	StopAutoRefreshFunc  func()
}

type ImageListPage struct {
	*tview.Flex      // Embeds the tview.Flex for its layout
	config           Config
	refreshTableFunc func() // Function to refresh this page's table
}

func (p *ImageListPage) RefreshTable() {
	images, err := p.config.UseCases.Query.ListAllImages(context.Background())

	if err != nil {
		p.config.DisplayStatus(fmt.Sprintf("Error fetching images: %v", err), 3*time.Second)
		return
	}

	// Now, wrap the UI update part in QueueUpdateDraw
	helper.PopulateImageTableUI(p.config.Table, images)
}

// NewImageListPage creates and returns the images page.
func NewImageListPage(cfg Config) *ImageListPage {
	// Initial refresh when page is created (optional, or call on switch)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Docker Images").SetTextAlign(tview.AlignCenter), 1, 0, false).
		AddItem(cfg.Table, 0, 1, true).
		AddItem(cfg.StatusBar, 1, 0, false).
		AddItem(cfg.ExitGuide, 1, 0, false)
	page := &ImageListPage{ // Create the ImageListPage instance
		Flex:   layout,
		config: cfg,
	}

	page.refreshTableFunc = func() {
		page.RefreshTable()
	}

	return page
}

// HandleInput is the page-specific input handler for the ImageListPage.
func (p *ImageListPage) HandleInput(
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

	// Image list specific keybindings (e.g., 'p' for pull, 'r' for remove, 'i' for inspect)
	switch event.Key() { // Handles tcell.Key constants
	// No Tab/Enter for toggle like containers, unless you add a filter
	default: // Handles rune-based keys
		if event.Rune() == 'q' { // 'q' to quit (global app exit from this page)
			p.config.App.Stop()
			return nil
		}

		// Actions that require an image to be selected
		if p.config.App.GetFocus() == p.config.Table { // Only proceed if table is focused
			row, _ := p.config.Table.GetSelection()
			if row < 1 { // Header row check
				p.config.DisplayStatus("No image selected (header row).", 2*time.Second)
				return nil
			}
			cell := p.config.Table.GetCell(row, 0)
			if cell == nil || cell.GetReference() == nil {
				p.config.DisplayStatus("Error: No image reference found.", 3*time.Second)
				return nil
			}
			image := cell.GetReference().(*domain.Image)

			switch event.Rune() {
			case 'p': // Pull Image (example action)
				p.config.DisplayStatus(fmt.Sprintf("Pulling image %s:%s...", image.Repository, image.Tag), 1*time.Second)
				go func() { // Perform Docker action in a goroutine
					err := fmt.Errorf("image pull not implemented yet for %s", image.Repository+":"+image.Tag)
					if err != nil {
						p.config.DisplayStatus(fmt.Sprintf("Failed to pull: %v", err), 5*time.Second)
					} else {
						p.config.DisplayStatus(fmt.Sprintf("Image %s pulled.", image.Repository+":"+image.Tag), 3*time.Second)
						p.refreshTableFunc() // Refresh table
					}
				}()
				return nil
			case 'r': // Remove Image (example action)
				p.config.DisplayStatus(fmt.Sprintf("Removing image %s:%s...", image.Repository, image.Tag), 1*time.Second)

				var identifier string
				// If the image has valid repository and tag, use that to untag it
				if image.Repository != "<none>" && image.Tag != "<none>" && image.Tag != "" {
					identifier = fmt.Sprintf("%s:%s", image.Repository, image.Tag)
				} else {
					identifier = image.ImageID
				}
				go func() { // Perform Docker action in a goroutine
					err := p.config.UseCases.Control.RemoveImage(context.Background(), identifier)

					p.config.App.QueueUpdateDraw(func() {
						if err != nil {
							p.config.DisplayStatus(fmt.Sprintf("Failed to remove: %v", err), 10*time.Second)
						} else {
							p.config.DisplayStatus(fmt.Sprintf("Image %s removed.", image.Repository+":"+image.Tag), 3*time.Second)
						}
						p.refreshTableFunc()
						p.config.App.SetFocus(p.config.Table)

					})

				}()
				return nil

			case 'i': // Inspect image - opens the modal
				p.config.DisplayStatus(fmt.Sprintf("Inspecting image %s...", image.ImageID[:12]), 1*time.Second)
				go func(selectedImage *domain.Image) {
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

					ImageID := selectedImage.ImageID
					ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()

					inspectRaw, err := p.config.UseCases.Query.ImageInspect(ctx, ImageID)
					if err != nil {
						p.config.App.QueueUpdateDraw(func() {
							p.config.DisplayStatus(fmt.Sprintf("Inspect error: %v", err), 7*time.Second)
							p.config.App.SetFocus(p.config.Table)
						})
						return
					}

					// Create the inspect modal content and add it as a new page
					inspectModalContent := inspect.CreateInspectModal(p.config.App, p.config.Pages, p.config.Table, inspectRaw, selectedImage.Repository, p.config.DisplayStatus)

					p.config.App.QueueUpdateDraw(func() {
						p.config.Pages.AddPage(PageInspectView, inspectModalContent, true, true)
						p.config.App.SetFocus(inspectModalContent.GetContentPrimitive())
					})
				}(image) // Pass the selected container to the goroutine
				return nil
			}

		}
	}
	return event
}

package images

import (
	// For Docker API calls
	"fmt"
	"time" // For time.Duration

	"docker-tui/internal/app"    // Your application use cases
	"docker-tui/internal/domain" // Your domain models

	// Your general UI helpers
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Constants for page names (should be defined in a central place)
const (
	PageMainMenu      = "main_menu"
	PageContainerList = "container_list"
	PageImageList     = "image_list" // NEW
	PageInspectView   = "inspect_view"
)

// Config struct to pass dependencies to the Images page
type Config struct {
	App                  *tview.Application
	Pages                *tview.Pages
	Table                *tview.Table // The actual image table
	StatusBar            *tview.TextView
	ExitGuide            *tview.TextView // This might change to a "Back to menu" guide
	UseCases             *app.ImageUseCases
	DisplayStatus        func(message string, duration time.Duration)
	StartAutoRefreshFunc func() // If image list also supports auto-refresh
	StopAutoRefreshFunc  func()
}

// ImageListPage represents the images view.
// It embeds a tview.Flex and stores its own internal state/dependencies.
type ImageListPage struct {
	*tview.Flex      // Embeds the tview.Flex for its layout
	config           Config
	refreshTableFunc func() // Function to refresh this page's table
}

// NewImageListPage creates and returns the images page.
func NewImageListPage(cfg Config, refreshTableFunc func()) *ImageListPage {
	// Initial refresh when page is created (optional, or call on switch)
	cfg.App.QueueUpdateDraw(func() {
		refreshTableFunc() // Populate image table
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Docker Images").SetTextAlign(tview.AlignCenter), 1, 0, false). // Simple Header
		AddItem(cfg.Table, 0, 1, true).                                                                     // Table is the primary focus of this view
		AddItem(cfg.StatusBar, 1, 0, false).
		AddItem(cfg.ExitGuide, 1, 0, false) // This guide might instruct to press 'm' for menu

	return &ImageListPage{
		Flex:             layout,
		config:           cfg,
		refreshTableFunc: refreshTableFunc,
	}
}

// HandleInput is the page-specific input handler for the ImageListPage.
func (p *ImageListPage) HandleInput(event *tcell.EventKey, mainMenuPage tview.Primitive, statusBar *tview.TextView, exitGuide *tview.TextView, refreshImageTable func(), displayTimedStatus func(message string, duration time.Duration), useCases *app.ImageUseCases) *tcell.EventKey { // Changed params
	// Handle keys to go back to the Main Menu
	if event.Rune() == 'm' { // 'm' for Menu
		p.config.App.QueueUpdateDraw(func() {
			p.config.Pages.SwitchToPage(PageMainMenu)
			p.config.App.SetFocus(mainMenuPage) // Set focus back to the menu list
		})
		return nil // Consume event
	}
	if event.Key() == tcell.KeyEscape {
		p.config.App.QueueUpdateDraw(func() {
			p.config.Pages.SwitchToPage(PageMainMenu)
			p.config.App.SetFocus(mainMenuPage)
		})
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
			image := cell.GetReference().(*domain.Image) // Assuming your image table stores domain.Image references

			switch event.Rune() {
			case 'p': // Pull Image (example action)
				p.config.DisplayStatus(fmt.Sprintf("Pulling image %s:%s...", image.Repository, image.Tag), 1*time.Second)
				go func() { // Perform Docker action in a goroutine
					// You'll need an ImageControlUseCase.PullImage method.
					// For now, this is a placeholder.
					// err := p.config.UseCases.Control.PullImage(context.Background(), image.Repository+":"+image.Tag)
					// Dummy error for example:
					err := fmt.Errorf("image pull not implemented yet for %s", image.Repository+":"+image.Tag)
					p.config.App.QueueUpdateDraw(func() { // Queue UI update
						if err != nil {
							p.config.DisplayStatus(fmt.Sprintf("Failed to pull: %v", err), 5*time.Second)
						} else {
							p.config.DisplayStatus(fmt.Sprintf("Image %s pulled.", image.Repository+":"+image.Tag), 3*time.Second)
							p.refreshTableFunc() // Refresh table
						}
					})
				}()
				return nil
			case 'r': // Remove Image (example action)
				p.config.DisplayStatus(fmt.Sprintf("Removing image %s:%s...", image.Repository, image.Tag), 1*time.Second)
				go func() { // Perform Docker action in a goroutine
					// You'll need an ImageControlUseCase.RemoveImage method.
					// err := p.config.UseCases.Control.RemoveImage(context.Background(), image.ImageID)
					// Dummy error for example:
					err := fmt.Errorf("image remove not implemented yet for %s", image.ImageID)
					p.config.App.QueueUpdateDraw(func() { // Queue UI update
						if err != nil {
							p.config.DisplayStatus(fmt.Sprintf("Failed to remove: %v", err), 5*time.Second)
						} else {
							p.config.DisplayStatus(fmt.Sprintf("Image %s removed.", image.Repository+":"+image.Tag), 3*time.Second)
							p.refreshTableFunc() // Refresh table
						}
					})
				}()
				return nil
				// You can add 'i' for inspect image here, similar to container inspect.
			} // End of inner switch event.Rune()
		} // End of if tuiApp.GetFocus() == p.config.Table
	} // End of default case for event.Key()
	return event // Propagate unhandled events
}

package helper

import (
	"bytes"
	"docker-tui/internal/domain"
	"encoding/json"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func PopulateContainerTableUI(
	table *tview.Table,
	containers []domain.Container,
) {
	// Clear old rows
	for row := table.GetRowCount() - 1; row >= 1; row-- {
		table.RemoveRow(row)
	}

	if len(containers) == 0 {
		// Center "No containers" message visually
		table.SetCell(1, 0, tview.NewTableCell("No containers found in this state.").
			SetSelectable(false).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetExpansion(7))
		for col := 1; col < table.GetColumnCount(); col++ { // Clear other cells in the row
			table.SetCell(1, col, tview.NewTableCell(""))
		}
	} else {
		for rowNum, container := range containers {
			rowIdx := rowNum + 1 // +1 for the header row
			table.SetCell(
				rowIdx, 0,
				tview.NewTableCell(container.ID[:12]).
					SetAlign(tview.AlignLeft).
					SetReference(&containers[rowNum]).
					SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan)))
			table.SetCell(rowIdx, 1, tview.NewTableCell(container.Image).SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 2,
				tview.NewTableCell(container.Command).
					SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 3,
				tview.NewTableCell(container.Created).
					SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 4,
				tview.NewTableCell(container.Status).
					SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 5,
				tview.NewTableCell(FormatContainerPorts(container.Ports)).
					SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 6,
				tview.NewTableCell(container.Name).
					SetAlign(tview.AlignLeft))
		}
	}
}

func PopulateImageTableUI(table *tview.Table, images []domain.Image) {
	for row := table.GetRowCount() - 1; row >= 1; row-- {
		table.RemoveRow(row)
	}

	if len(images) == 0 {
		// Center "No images" message visually
		table.SetCell(1, 0, tview.NewTableCell("No images found.").
			SetSelectable(false).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetExpansion(7))
		for col := 1; col < table.GetColumnCount(); col++ { // Clear other cells in the row
			table.SetCell(1, col, tview.NewTableCell(""))
		}
	} else {
		for rowNum, image := range images {
			rowIdx := rowNum + 1 // +1 for the header row
			table.SetCell(rowIdx, 0, tview.NewTableCell(image.Repository).
				SetReference(&image).
				SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 1,
				tview.NewTableCell(image.Tag).
					SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 2,
				tview.NewTableCell(image.ImageID[:12]).
					SetAlign(tview.AlignLeft).
					SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan)))
			table.SetCell(
				rowIdx, 3,
				tview.NewTableCell(image.Created).
					SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 4,
				tview.NewTableCell(image.Size).
					SetAlign(tview.AlignLeft))
		}
	}
}

func PopulateVolumeTableUI(table *tview.Table, volumes []domain.Volume) {
	for row := table.GetRowCount() - 1; row >= 1; row-- {
		table.RemoveRow(row)
	}

	if len(volumes) == 0 {
		// Center "No images" message visually
		table.SetCell(1, 0, tview.NewTableCell("No volume found.").
			SetSelectable(false).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetExpansion(7))
		for col := 1; col < table.GetColumnCount(); col++ { // Clear other cells in the row
			table.SetCell(1, col, tview.NewTableCell(""))
		}
	} else {
		for rowNum, volume := range volumes {
			rowIdx := rowNum + 1 // +1 for the header row
			table.SetCell(rowIdx, 0, tview.NewTableCell(volume.Name[:12]).
				SetReference(&volume).
				SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 1,
				tview.NewTableCell(volume.Driver).
					SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 2,
				tview.NewTableCell(volume.Mountpoint[24:]).
					SetAlign(tview.AlignLeft))
		}
	}
}

func PopulateNetworkTableUI(table *tview.Table, networks []domain.Network) {
	for row := table.GetRowCount() - 1; row >= 1; row-- {
		table.RemoveRow(row)
	}

	if len(networks) == 0 {
		// Center "No images" message visually
		table.SetCell(1, 0, tview.NewTableCell("No network found.").
			SetSelectable(false).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetExpansion(7))
		for col := 1; col < table.GetColumnCount(); col++ { // Clear other cells in the row
			table.SetCell(1, col, tview.NewTableCell(""))
		}
	} else {
		for rowNum, network := range networks {
			rowIdx := rowNum + 1 // +1 for the header row
			table.SetCell(rowIdx, 0, tview.NewTableCell(network.NetworkID).
				SetReference(&network).
				SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 1,
				tview.NewTableCell(network.Name).
					SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 2,
				tview.NewTableCell(network.Driver).
					SetAlign(tview.AlignLeft))
			table.SetCell(
				rowIdx, 3,
				tview.NewTableCell(network.Scope).
					SetAlign(tview.AlignLeft))
		}
	}
}

func ContainerTableFormat() *tview.Table {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, true)
	table.SetFixed(1, 0)

	headers := []string{
		"CONTAINER ID", "IMAGE", "COMMAND", "CREATED", "STATUS", "PORTS", "NAMES",
	}

	table.SetTitle("Docker Container - CLI Based").SetTitleAlign(tview.AlignCenter)

	for col, header := range headers {
		table.SetCell(0, col,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
	}

	return table
}

func ImageTableFormat() *tview.Table {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, true)
	table.SetFixed(1, 0)

	headers := []string{
		"REPOSITORY",
		"TAG",
		"IMAGE ID",
		"CREATED",
		"SIZE",
	}

	table.SetTitle("Docker Container - CLI Based").SetTitleAlign(tview.AlignCenter)

	for col, header := range headers {
		table.SetCell(0, col,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	return table
}

func NetworkTableFormat() *tview.Table {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, true)
	table.SetFixed(1, 0)

	headers := []string{
		"NETWORK ID",
		"NAME",
		"DRIVER",
		"SCOPE",
	}

	table.SetTitle("Docker Container - CLI Based").SetTitleAlign(tview.AlignCenter)

	for col, header := range headers {
		table.SetCell(0, col,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	return table
}

func VolumeTableFormat() *tview.Table {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, true)
	table.SetFixed(1, 0)

	headers := []string{
		"NAME",
		"DRIVER",
		"MOUNTPOINT",
	}

	table.SetTitle("Docker Container - CLI Based").SetTitleAlign(tview.AlignCenter)

	for col, header := range headers {
		table.SetCell(0, col,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	return table
}

func PrettyJson(data string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(data), "", "  "); err != nil {
		// Fallback to raw JSON if pretty-printing fails
		return string(data), nil
	}
	return prettyJSON.String(), nil
}

func CloseModal(app *tview.Application, pages *tview.Pages, pageName string, focusPrimitive tview.Primitive) {
	pages.RemovePage(pageName)
	if focusPrimitive != nil {
		app.SetFocus(focusPrimitive)
	}
}

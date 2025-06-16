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

func TableFormat() *tview.Table {
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

func PrettyJson(data string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(data), "", "  "); err != nil {
		// Fallback to raw JSON if pretty-printing fails
		return string(data), nil
	}
	return prettyJSON.String(), nil
}

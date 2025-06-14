package helper

import (
	"context"
	"docker-tui/internal/app"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func RefreshTable(
	ctx context.Context,
	containerState string,
	toggleState *tview.TextView,
	table *tview.Table,
	statusBar *tview.TextView,
	app *app.ContainerUseCases,
) {

	UpdateToggleText(containerState, toggleState)
	containers, err := app.Query.ListContainersByState(ctx, containerState)
	if err != nil {
		statusBar.SetText(fmt.Sprintf("Error: %v", err))
		return
	}

	// Clear old rows
	for row := table.GetRowCount() - 1; row >= 1; row-- {
		table.RemoveRow(row)
	}

	for rowNum, container := range containers {
		rowIdx := rowNum + 1
		table.SetCell(rowIdx, 0, tview.NewTableCell(container.ID[:12]).SetAlign(tview.AlignLeft).SetReference(&containers[rowNum]))
		table.SetCell(rowIdx, 1, tview.NewTableCell(container.Image).SetAlign(tview.AlignLeft))
		table.SetCell(rowIdx, 2, tview.NewTableCell(container.Command).SetAlign(tview.AlignLeft))
		table.SetCell(rowIdx, 3, tview.NewTableCell(container.Created).SetAlign(tview.AlignLeft))
		table.SetCell(rowIdx, 4, tview.NewTableCell(container.Status).SetAlign(tview.AlignLeft))
		table.SetCell(rowIdx, 5, tview.NewTableCell(FormatContainerPorts(container.Ports)).SetAlign(tview.AlignLeft))
		table.SetCell(rowIdx, 6, tview.NewTableCell(container.Name).SetAlign(tview.AlignLeft))
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

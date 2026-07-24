package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	accentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	headerStyle   = accentStyle.Bold(true)
	faintStyle    = lipgloss.NewStyle().Faint(true)
	panelStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("6")).Foreground(lipgloss.Color("0"))
)

func (m Model) View() string {
	if m.width == 0 {
		return "loading herdr..."
	}

	items := m.visibleItems()
	resultCount := fmt.Sprintf("%d results", len(items))
	title := headerStyle.Render("herdr")
	space := max(1, m.width-lipgloss.Width(title)-lipgloss.Width(resultCount))
	header := title + strings.Repeat(" ", space) + faintStyle.Render(resultCount)
	search := accentStyle.Bold(true).Render("/") + " " + m.input.View()
	scopes := make([]string, 0, 4)
	for scope := ScopeAll; scope <= ScopeTabs; scope++ {
		label := scope.String()
		if scope == m.scope {
			label = headerStyle.Render("[" + label + "]")
		} else {
			label = faintStyle.Render(" " + label + " ")
		}
		scopes = append(scopes, label)
	}

	bodyHeight := max(1, m.height-4)
	resultsHeight := max(1, bodyHeight*35/100)
	previewHeight := bodyHeight - resultsHeight - 1
	body := m.resultsPanel(m.width, resultsHeight)
	if previewHeight > 0 {
		body += "\n" + m.previewPanel(m.width, previewHeight)
	}
	footer := faintStyle.Render(ansi.Truncate("up/down navigate   tab/shift+tab scope   enter focus   esc clear/close", m.width, ""))
	return strings.Join([]string{header, search, strings.Join(scopes, ""), body, footer}, "\n")
}

func (m Model) resultsPanel(width, height int) string {
	innerWidth, innerHeight := max(1, width-2), max(1, height-2)
	heading := faintStyle.Render("results")
	content := heading
	if innerHeight > 1 {
		content += "\n" + m.resultsView(innerWidth, innerHeight-1)
	}
	return panelStyle.Width(innerWidth).Height(innerHeight).Render(content)
}

func (m Model) resultsView(width, height int) string {
	items := m.visibleItems()
	if len(items) == 0 {
		return faintStyle.Render("  no matching results")
	}
	start := 0
	if m.selected >= height {
		start = m.selected - height + 1
	}
	end := min(len(items), start+height)
	rows := make([]string, 0, end-start)
	for index := start; index < end; index++ {
		item := items[index]
		kind := fmt.Sprintf("%-9s", strings.TrimSuffix(item.Kind.String(), "s"))
		row := "  " + faintStyle.Render(kind) + " " + item.Label
		if item.Detail != "" {
			row += faintStyle.Render("  " + item.Detail)
		}
		row = ansi.Truncate(row, width, "…")
		if index == m.selected {
			plainWidth := ansi.StringWidth(row)
			if plainWidth < width {
				row += strings.Repeat(" ", width-plainWidth)
			}
			row = selectedStyle.Render(">" + strings.TrimPrefix(row, " "))
		}
		rows = append(rows, row)
	}
	return strings.Join(rows, "\n")
}

func (m Model) previewPanel(width, height int) string {
	content := m.preview
	if m.loading || m.status != "" {
		content = m.status
	}
	innerWidth := max(1, width-2)
	innerHeight := max(1, height-2)
	previewHeight := max(0, innerHeight-2)
	preview := viewport.New(innerWidth, previewHeight)
	preview.SetContent(content)
	title := "preview"
	path := ""
	if items := m.visibleItems(); len(items) > 0 && m.selected < len(items) {
		path = items[m.selected].PreviewPath()
	}
	heading := faintStyle.Render(title)
	if path != "" {
		heading += "  " + ansi.Truncate(path, max(1, innerWidth-10), "…")
	}
	return panelStyle.Width(innerWidth).Height(innerHeight).Render(heading + "\n\n" + preview.View())
}

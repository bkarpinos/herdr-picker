package main

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestViewStaysWithinTerminal(t *testing.T) {
	for _, size := range []struct{ width, height int }{{120, 20}, {60, 20}} {
		m := NewModel(nil)
		m.width, m.height = size.width, size.height
		m.loading = false
		m.items = []Item{{Kind: ScopeAgents, ID: "p1", Label: "opencode", WorkspaceLabel: "API", TabLabel: "Dev", PaneID: "p1"}}
		m.preview = "one\ntwo\nthree"
		view := m.View()
		if got := lipgloss.Width(view); got > size.width {
			t.Errorf("width at %d columns = %d", size.width, got)
		}
		if got := lipgloss.Height(view); got > size.height {
			t.Errorf("height at %d rows = %d", size.height, got)
		}
	}
}

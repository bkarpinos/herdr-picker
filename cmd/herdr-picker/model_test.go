package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestVisibleItemsRanksLabelBeforeDetail(t *testing.T) {
	m := NewModel(nil)
	m.items = []Item{
		{Kind: ScopeTabs, Label: "backend", Detail: "api"},
		{Kind: ScopeTabs, Label: "api", Detail: "backend"},
	}
	m.input.SetValue("api")
	items := m.visibleItems()
	if items[0].Kind != ScopeTabs {
		t.Fatalf("first match = %s, want tab", items[0].Kind)
	}
}

func TestVisibleItemsGroupsKinds(t *testing.T) {
	m := NewModel(nil)
	m.items = []Item{
		{Kind: ScopeTabs, Label: "api tab"},
		{Kind: ScopeAgents, Label: "api agent"},
		{Kind: ScopeWorkspaces, Label: "api workspace"},
	}
	m.input.SetValue("api")
	items := m.visibleItems()
	want := []Scope{ScopeWorkspaces, ScopeAgents, ScopeTabs}
	for index, kind := range want {
		if items[index].Kind != kind {
			t.Fatalf("item %d kind = %s, want %s", index, items[index].Kind, kind)
		}
	}
}

func TestScopeFilteringKeepsQuery(t *testing.T) {
	m := NewModel(nil)
	m.items = []Item{{Kind: ScopeTabs, Label: "api"}, {Kind: ScopeAgents, Label: "api-agent"}}
	m.input.SetValue("api")
	m.setScope(ScopeAgents)
	items := m.visibleItems()
	if m.input.Value() != "api" || len(items) != 1 || items[0].Kind != ScopeAgents {
		t.Fatalf("scope filtering lost query or returned wrong items")
	}
}

func TestScopeChangePreservesSelectedItem(t *testing.T) {
	m := NewModel(nil)
	m.items = []Item{{Kind: ScopeTabs, ID: "tab", Label: "api"}, {Kind: ScopeAgents, ID: "agent", Label: "api-agent"}}
	m.selected = 1
	m.setScope(ScopeAgents)
	items := m.visibleItems()
	if m.selected != 0 || items[m.selected].ID != "agent" {
		t.Fatalf("selected item changed after scope transition")
	}
}

func TestTabKeysCycleScopes(t *testing.T) {
	m := NewModel(nil)
	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
	if got := updated.(Model).scope; got != ScopeWorkspaces {
		t.Fatalf("Tab scope = %s, want Workspaces", got)
	}
	updated, _ = updated.(Model).handleKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	if got := updated.(Model).scope; got != ScopeAll {
		t.Fatalf("Shift+Tab scope = %s, want All", got)
	}
	updated, _ = updated.(Model).handleKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	if got := updated.(Model).scope; got != ScopeTabs {
		t.Fatalf("wrapped Shift+Tab scope = %s, want Tabs", got)
	}
}

func TestPreviewTickSuppressesStaleRequest(t *testing.T) {
	runner := &testRunner{responses: map[string][]byte{
		"pane read p1 --source visible --ansi": []byte("content"),
	}}
	m := NewModel(runner)
	m.items = []Item{{Kind: ScopeAgents, ID: "p1", PaneID: "p1"}}
	m.previewGeneration = 2

	updated, command := m.Update(previewTickMsg{generation: 1, item: m.items[0]})
	if command != nil || runner.callCount() != 0 {
		t.Fatal("stale preview tick started a request")
	}

	updated, command = updated.(Model).Update(previewTickMsg{generation: 2, item: m.items[0]})
	if command == nil {
		t.Fatal("current preview tick did not start a request")
	}
	message := command().(previewMsg)
	if message.err != nil || message.content != "content" {
		t.Fatalf("preview message = %#v", message)
	}
	if got := runner.lastCall(); got != "pane read p1 --source visible --ansi" {
		t.Fatalf("last command = %q", got)
	}
}

func TestStalePreviewResponseIsIgnored(t *testing.T) {
	m := NewModel(nil)
	m.preview = "current"
	m.previewGeneration = 2
	updated, _ := m.Update(previewMsg{generation: 1, content: "stale"})
	if got := updated.(Model).preview; got != "current" {
		t.Fatalf("preview = %q, want current", got)
	}
}

func TestFocusCommandForEachItemKind(t *testing.T) {
	tests := []struct {
		item Item
		want string
	}{
		{Item{Kind: ScopeWorkspaces, ID: "w1"}, "workspace focus w1"},
		{Item{Kind: ScopeTabs, ID: "t1"}, "tab focus t1"},
		{Item{Kind: ScopeAgents, ID: "p1"}, "agent focus p1"},
	}
	for _, test := range tests {
		t.Run(test.item.Kind.String(), func(t *testing.T) {
			runner := &testRunner{responses: map[string][]byte{test.want: {}}}
			m := NewModel(runner)
			m.items = []Item{test.item}
			message := m.focusCmd()().(focusMsg)
			if message.err != nil {
				t.Fatal(message.err)
			}
			if got := runner.lastCall(); got != test.want {
				t.Fatalf("focus command = %q, want %q", got, test.want)
			}
		})
	}
}

func TestEscapeClearsQueryBeforeQuitting(t *testing.T) {
	m := NewModel(nil)
	m.input.SetValue("api")
	updated, command := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if command != nil {
		t.Fatal("escape with a query returned an unexpected command")
	}
	if got := updated.(Model).input.Value(); got != "" {
		t.Fatalf("query = %q, want empty", got)
	}
}

func TestResultsViewKeepsSelectionVisible(t *testing.T) {
	m := NewModel(nil)
	for index := 0; index < 10; index++ {
		m.items = append(m.items, Item{Kind: ScopeTabs, ID: string(rune('a' + index)), Label: string(rune('a' + index))})
	}
	m.selected = 9
	view := m.resultsView(40, 3)
	if lines := strings.Count(view, "\n") + 1; lines != 3 {
		t.Fatalf("rendered lines = %d, want 3", lines)
	}
	if !strings.Contains(view, "> tab") || !strings.Contains(view, "j") {
		t.Fatalf("selected item is not visible:\n%s", view)
	}
}

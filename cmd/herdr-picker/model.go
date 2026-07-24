package main

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type dataMsg struct {
	items []Item
	err   error
}
type previewMsg struct {
	generation int
	content    string
	err        error
}
type previewTickMsg struct {
	generation int
	item       Item
}
type focusMsg struct{ err error }

type Model struct {
	runner            Runner
	input             textinput.Model
	items             []Item
	scope             Scope
	selected          int
	preview           string
	status            string
	loading           bool
	previewGeneration int
	width, height     int
}

func NewModel(runner Runner) Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "search"
	input.Focus()
	input.CharLimit = 256
	return Model{runner: runner, input: input, loading: true, status: "loading herdr entities..."}
}

func (m Model) Init() tea.Cmd { return m.loadCmd() }

func (m Model) loadCmd() tea.Cmd {
	return func() tea.Msg {
		items, err := loadItems(context.Background(), m.runner)
		return dataMsg{items: items, err: err}
	}
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = message.Width, message.Height
		m.input.Width = max(10, message.Width-4)
		return m, nil
	case dataMsg:
		m.loading = false
		if message.err != nil {
			m.status = message.err.Error()
			return m, nil
		}
		m.items, m.status = message.items, ""
		return m, m.schedulePreview()
	case previewMsg:
		if message.generation != m.previewGeneration {
			return m, nil
		}
		if message.err != nil {
			m.preview = "preview unavailable: " + message.err.Error()
		} else {
			m.preview = message.content
		}
		return m, nil
	case previewTickMsg:
		if message.generation != m.previewGeneration {
			return m, nil
		}
		m.preview = "loading preview..."
		return m, m.previewCmd(message.generation, message.item)
	case focusMsg:
		if message.err != nil {
			m.status = "focus failed: " + message.err.Error()
			return m, nil
		}
		return m, tea.Quit
	case tea.KeyMsg:
		return m.handleKey(message)
	}

	previousQuery := m.input.Value()
	selectedID := m.selectedID()
	var command tea.Cmd
	m.input, command = m.input.Update(message)
	if m.input.Value() != previousQuery {
		m.restoreSelection(selectedID)
		return m, tea.Batch(command, m.schedulePreview())
	}
	return m, command
}

func (m Model) handleKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.input.Value() != "" {
			selectedID := m.selectedID()
			m.input.SetValue("")
			m.restoreSelection(selectedID)
			return m, m.schedulePreview()
		}
		return m, tea.Quit
	case "up", "ctrl+p":
		return m.move(-1)
	case "down", "ctrl+n":
		return m.move(1)
	case "tab":
		m.setScope(Scope((int(m.scope) + 1) % 4))
		return m, m.schedulePreview()
	case "shift+tab":
		m.setScope(Scope((int(m.scope) + 3) % 4))
		return m, m.schedulePreview()
	case "enter":
		return m, m.focusCmd()
	default:
		previousQuery := m.input.Value()
		selectedID := m.selectedID()
		var command tea.Cmd
		m.input, command = m.input.Update(key)
		if m.input.Value() != previousQuery {
			m.restoreSelection(selectedID)
			return m, tea.Batch(command, m.schedulePreview())
		}
		return m, command
	}
}

func (m *Model) setScope(scope Scope) {
	selectedID := m.selectedID()
	m.scope = scope
	m.restoreSelection(selectedID)
}

func (m Model) move(delta int) (tea.Model, tea.Cmd) {
	items := m.visibleItems()
	if len(items) == 0 {
		return m, nil
	}
	m.selected = (m.selected + delta + len(items)) % len(items)
	return m, m.schedulePreview()
}

func (m Model) selectedID() string {
	items := m.visibleItems()
	if m.selected < len(items) {
		return items[m.selected].Kind.String() + ":" + items[m.selected].ID
	}
	return ""
}

func (m *Model) restoreSelection(selectedID string) {
	items := m.visibleItems()
	for index, item := range items {
		if item.Kind.String()+":"+item.ID == selectedID {
			m.selected = index
			return
		}
	}
	m.selected = 0
}

func (m Model) visibleItems() []Item {
	query := strings.ToLower(strings.TrimSpace(m.input.Value()))
	items := make([]Item, 0, len(m.items))
	for _, item := range m.items {
		if (m.scope == ScopeAll || item.Kind == m.scope) && (query == "" || strings.Contains(item.SearchText(), query)) {
			items = append(items, item)
		}
	}
	if query != "" {
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].Kind != items[j].Kind {
				return items[i].Kind < items[j].Kind
			}
			return matchScore(items[i], query) < matchScore(items[j], query)
		})
	}
	return items
}

func matchScore(item Item, query string) int {
	label := strings.ToLower(item.Label)
	switch {
	case label == query:
		return 0
	case strings.HasPrefix(label, query):
		return 1
	case strings.Contains(label, query):
		return 2
	case strings.Contains(strings.ToLower(item.TerminalTitle), query):
		return 3
	case strings.Contains(strings.ToLower(item.Detail), query):
		return 4
	default:
		return 5
	}
}

func (m *Model) schedulePreview() tea.Cmd {
	m.previewGeneration++
	generation := m.previewGeneration
	items := m.visibleItems()
	if len(items) == 0 {
		m.preview = "no matching result."
		return nil
	}
	item := items[m.selected]
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return previewTickMsg{generation: generation, item: item}
	})
}

func (m Model) previewCmd(generation int, item Item) tea.Cmd {
	return func() tea.Msg {
		paneID, err := resolvePreviewPane(context.Background(), m.runner, item)
		if err != nil {
			return previewMsg{generation: generation, err: err}
		}
		output, err := m.runner.Run(context.Background(), "pane", "read", paneID, "--source", "visible", "--ansi")
		return previewMsg{generation: generation, content: string(output), err: err}
	}
}

func (m Model) focusCmd() tea.Cmd {
	items := m.visibleItems()
	if len(items) == 0 {
		return nil
	}
	item := items[m.selected]
	return func() tea.Msg {
		var args []string
		switch item.Kind {
		case ScopeWorkspaces:
			args = []string{"workspace", "focus", item.ID}
		case ScopeTabs:
			args = []string{"tab", "focus", item.ID}
		case ScopeAgents:
			args = []string{"agent", "focus", item.ID}
		}
		_, err := m.runner.Run(context.Background(), args...)
		return focusMsg{err: err}
	}
}

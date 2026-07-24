package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
)

type workspace struct {
	ID          string `json:"workspace_id"`
	Label       string `json:"label"`
	ActiveTabID string `json:"active_tab_id"`
}

type tab struct {
	ID          string `json:"tab_id"`
	Label       string `json:"label"`
	WorkspaceID string `json:"workspace_id"`
}

type pane struct {
	ID            string `json:"pane_id"`
	WorkspaceID   string `json:"workspace_id"`
	TabID         string `json:"tab_id"`
	CWD           string `json:"foreground_cwd"`
	Agent         string `json:"agent"`
	TerminalTitle string `json:"terminal_title"`
}

type agent struct {
	Name          string `json:"agent"`
	PaneID        string `json:"pane_id"`
	WorkspaceID   string `json:"workspace_id"`
	TabID         string `json:"tab_id"`
	CWD           string `json:"foreground_cwd"`
	TerminalTitle string `json:"terminal_title"`
}

type lists struct {
	workspaces []workspace
	tabs       []tab
	panes      []pane
	agents     []agent
}

func loadItems(ctx context.Context, runner Runner) ([]Item, error) {
	type response struct {
		kind string
		data []byte
		err  error
	}
	requests := []struct {
		kind string
		args []string
	}{
		{"workspaces", []string{"workspace", "list"}},
		{"tabs", []string{"tab", "list"}},
		{"panes", []string{"pane", "list"}},
		{"agents", []string{"agent", "list"}},
	}
	responses := make(chan response, len(requests))
	for _, request := range requests {
		go func(kind string, args []string) {
			data, err := runner.Run(ctx, args...)
			responses <- response{kind: kind, data: data, err: err}
		}(request.kind, request.args)
	}

	var result lists
	for range requests {
		response := <-responses
		if response.err != nil {
			return nil, fmt.Errorf("herdr %s list: %w", response.kind, response.err)
		}
		if err := decodeList(response.data, response.kind, &result); err != nil {
			return nil, err
		}
	}
	return normalize(result), nil
}

func decodeList(data []byte, kind string, target *lists) error {
	var envelope struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	switch kind {
	case "workspaces":
		var value struct {
			Workspaces []workspace `json:"workspaces"`
		}
		if err := json.Unmarshal(envelope.Result, &value); err != nil {
			return err
		}
		target.workspaces = value.Workspaces
	case "tabs":
		var value struct {
			Tabs []tab `json:"tabs"`
		}
		if err := json.Unmarshal(envelope.Result, &value); err != nil {
			return err
		}
		target.tabs = value.Tabs
	case "panes":
		var value struct {
			Panes []pane `json:"panes"`
		}
		if err := json.Unmarshal(envelope.Result, &value); err != nil {
			return err
		}
		target.panes = value.Panes
	case "agents":
		var value struct {
			Agents []agent `json:"agents"`
		}
		if err := json.Unmarshal(envelope.Result, &value); err != nil {
			return err
		}
		target.agents = value.Agents
	}
	return nil
}

func normalize(data lists) []Item {
	workspaces := make(map[string]workspace, len(data.workspaces))
	for _, value := range data.workspaces {
		workspaces[value.ID] = value
	}
	tabs := make(map[string]tab, len(data.tabs))
	for _, value := range data.tabs {
		tabs[value.ID] = value
	}
	firstPaneByTab := make(map[string]string, len(data.tabs))
	for _, value := range data.panes {
		if firstPaneByTab[value.TabID] == "" {
			firstPaneByTab[value.TabID] = value.ID
		}
	}
	agentTabs := make(map[string]bool, len(data.agents))
	for _, value := range data.agents {
		agentTabs[value.TabID] = true
	}

	items := make([]Item, 0, len(data.workspaces)+len(data.tabs)+len(data.agents))
	for _, value := range data.workspaces {
		activeTab := tabs[value.ActiveTabID]
		items = append(items, Item{Kind: ScopeWorkspaces, ID: value.ID, Label: value.Label, Detail: value.ID, WorkspaceID: value.ID, WorkspaceLabel: value.Label, TabID: value.ActiveTabID, TabLabel: activeTab.Label, PaneID: firstPaneByTab[value.ActiveTabID], ResolveLayout: true})
	}
	for _, value := range data.agents {
		items = append(items, Item{Kind: ScopeAgents, ID: value.PaneID, Label: value.Name, Detail: filepath.Base(value.CWD), WorkspaceID: value.WorkspaceID, WorkspaceLabel: workspaces[value.WorkspaceID].Label, TabID: value.TabID, TabLabel: tabs[value.TabID].Label, PaneID: value.PaneID, TerminalTitle: value.TerminalTitle})
	}
	for _, value := range data.tabs {
		if agentTabs[value.ID] {
			continue
		}
		workspaceLabel := workspaces[value.WorkspaceID].Label
		items = append(items, Item{Kind: ScopeTabs, ID: value.ID, Label: value.Label, Detail: workspaceLabel, WorkspaceID: value.WorkspaceID, WorkspaceLabel: workspaceLabel, TabID: value.ID, TabLabel: value.Label, PaneID: firstPaneByTab[value.ID], ResolveLayout: true})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Kind != items[j].Kind {
			return items[i].Kind < items[j].Kind
		}
		return items[i].Title() < items[j].Title()
	})
	return items
}

func resolvePreviewPane(ctx context.Context, runner Runner, item Item) (string, error) {
	if item.PaneID != "" && !item.ResolveLayout {
		return item.PaneID, nil
	}
	seed := item.PaneID
	if seed == "" {
		return "", fmt.Errorf("no pane available for %s", item.Title())
	}
	data, err := runner.Run(ctx, "pane", "layout", "--pane", seed)
	if err != nil {
		return "", err
	}
	var response struct {
		Result struct {
			Layout struct {
				FocusedPaneID string `json:"focused_pane_id"`
			} `json:"layout"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return "", err
	}
	if response.Result.Layout.FocusedPaneID == "" {
		return "", fmt.Errorf("layout has no focused pane")
	}
	return response.Result.Layout.FocusedPaneID, nil
}

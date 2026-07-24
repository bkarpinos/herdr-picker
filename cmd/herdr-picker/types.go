package main

import "strings"

type Scope int

const (
	ScopeAll Scope = iota
	ScopeWorkspaces
	ScopeAgents
	ScopeTabs
)

func (s Scope) String() string {
	return [...]string{"all", "workspaces", "agents", "tabs"}[s]
}

type Item struct {
	Kind           Scope
	ID             string
	Label          string
	Detail         string
	WorkspaceID    string
	WorkspaceLabel string
	TabID          string
	TabLabel       string
	PaneID         string
	TerminalTitle  string
	ResolveLayout  bool
}

func (i Item) SearchText() string {
	return strings.ToLower(strings.Join([]string{i.Label, i.Detail, i.WorkspaceLabel, i.TabLabel, i.TerminalTitle, i.WorkspaceID, i.TabID, i.PaneID}, " "))
}

func (i Item) Title() string {
	return strings.ToLower(i.Kind.String()[:len(i.Kind.String())-1]) + ": " + i.Label
}

func (i Item) PreviewPath() string {
	workspace := i.WorkspaceLabel
	if workspace == "" {
		workspace = i.WorkspaceID
	}
	tab := i.TabLabel
	if tab == "" {
		tab = i.TabID
	}
	parts := make([]string, 0, 3)
	for _, part := range []string{workspace, tab, i.PaneID} {
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return "preview"
	}
	return strings.Join(parts, " / ")
}

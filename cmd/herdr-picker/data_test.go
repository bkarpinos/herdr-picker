package main

import (
	"context"
	"testing"
)

func TestLoadItemsNormalizesHierarchy(t *testing.T) {
	runner := &testRunner{responses: map[string][]byte{
		"workspace list": []byte(`{"result":{"workspaces":[{"workspace_id":"w1","label":"API","active_tab_id":"t1"}]}}`),
		"tab list":       []byte(`{"result":{"tabs":[{"tab_id":"t1","label":"Server","workspace_id":"w1"},{"tab_id":"t2","label":"Logs","workspace_id":"w1"}]}}`),
		"pane list":      []byte(`{"result":{"panes":[{"pane_id":"p1","workspace_id":"w1","tab_id":"t1"},{"pane_id":"p2","workspace_id":"w1","tab_id":"t2"}]}}`),
		"agent list":     []byte(`{"result":{"agents":[{"agent":"opencode","pane_id":"p1","workspace_id":"w1","tab_id":"t1","foreground_cwd":"/work/api"}]}}`),
	}}

	items, err := loadItems(context.Background(), runner)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}
	if runner.callCount() != 4 {
		t.Fatalf("runner calls = %d, want 4", runner.callCount())
	}

	wantKinds := []Scope{ScopeWorkspaces, ScopeAgents, ScopeTabs}
	for index, want := range wantKinds {
		if items[index].Kind != want {
			t.Fatalf("item %d kind = %s, want %s", index, items[index].Kind, want)
		}
	}
	if items[2].ID != "t2" {
		t.Fatalf("tab item = %q, want non-agent tab t2", items[2].ID)
	}
	if got := items[1].PreviewPath(); got != "API / Server / p1" {
		t.Fatalf("preview path = %q", got)
	}
}

func TestResolvePreviewPane(t *testing.T) {
	t.Run("known pane", func(t *testing.T) {
		runner := &testRunner{}
		paneID, err := resolvePreviewPane(context.Background(), runner, Item{PaneID: "p1"})
		if err != nil || paneID != "p1" {
			t.Fatalf("resolvePreviewPane() = %q, %v", paneID, err)
		}
		if runner.callCount() != 0 {
			t.Fatal("known pane should not load layout")
		}
	})

	t.Run("focused layout pane", func(t *testing.T) {
		runner := &testRunner{responses: map[string][]byte{
			"pane layout --pane p1": []byte(`{"result":{"layout":{"focused_pane_id":"p2"}}}`),
		}}
		paneID, err := resolvePreviewPane(context.Background(), runner, Item{PaneID: "p1", ResolveLayout: true})
		if err != nil || paneID != "p2" {
			t.Fatalf("resolvePreviewPane() = %q, %v", paneID, err)
		}
	})
}

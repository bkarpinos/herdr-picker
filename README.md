# herdr picker

herdr picker is a popup for searching workspaces, agents, and tabs. It previews the selected entity's terminal content and focuses it. Tabs that host detected agents are represented by their agent row instead of appearing twice.

## Requirements

- herdr 0.7.5 or newer
- Go 1.24 or newer when building locally
- Linux or macOS

## Local Installation

Build and link the working tree:

```sh
go test ./...
go build -o bin/herdr-picker ./cmd/herdr-picker
herdr plugin link .
```

`herdr plugin link` does not run the manifest's build command, so rebuild the binary after local code changes.
After changing `herdr-plugin.toml`, run `herdr plugin unlink herdr.picker` and link it again so herdr reloads the manifest.

Confirm that herdr registered the plugin:

```sh
herdr plugin list
herdr plugin action list --plugin herdr.picker
```

Open the picker without configuring a keybinding:

```sh
herdr plugin action invoke open --plugin herdr.picker
```

## Keybinding

Add this entry to `~/.config/herdr/config.toml`:

```toml
[[keys.command]]
key = "prefix+f"
type = "plugin_action"
command = "herdr.picker.open"
description = "open herdr picker"
```

Apply it to a running server:

```sh
herdr server reload-config
```

## Controls

| Key | Action |
| --- | --- |
| `Up`, `Ctrl+P` | Select previous result |
| `Down`, `Ctrl+N` | Select next result |
| `Tab` | Select next scope |
| `Shift+Tab` | Select previous scope |
| `Enter` | Focus the selection and close |
| `Esc` | Clear the query, or close when empty |
| `Ctrl+C` | Close |

## Manual Test

1. Run `herdr --version` and upgrade with `herdr update` if it is older than 0.7.5.
2. Run the local installation commands above from this repository.
3. Ensure the active herdr session has multiple workspaces and tabs, including at least one detected agent.
4. Invoke `herdr plugin action invoke open --plugin herdr.picker` from a herdr pane.
5. Confirm the popup is approximately 90% wide and 80% high and the search field is focused.
6. Type parts of a workspace label, tab label, agent name, and working directory; confirm results remain grouped as Workspaces, Agents, then Tabs.
7. Navigate quickly with arrows and `Ctrl+N`/`Ctrl+P`; confirm the selection remains visible and the preview settles on the final selection without flicker.
8. Confirm the preview header shows `workspace / tab / pane` and the body preserves the styling from `herdr pane read <pane-id> --source visible --ansi`.
9. Confirm tabs containing detected agents do not also appear as tab results.
10. Cycle scopes with `Tab` and `Shift+Tab`; confirm the query remains and the selected item is preserved when present.
11. Press `Esc` with a query to clear it, then press `Esc` again to close.
12. Reopen and press `Enter` on workspace, tab, and agent results; confirm each target receives focus and the popup closes.
13. Resize the outer terminal; confirm the vertically stacked results and preview remain bounded.

To remove the local link:

```sh
herdr plugin unlink herdr.picker
```

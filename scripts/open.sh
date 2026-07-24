#!/bin/sh
set -eu

"${HERDR_BIN_PATH:-herdr}" plugin pane open --plugin herdr.picker --entrypoint picker

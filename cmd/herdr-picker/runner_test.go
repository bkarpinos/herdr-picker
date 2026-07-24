package main

import (
	"context"
	"strings"
	"testing"
)

func TestCommandRunnerIncludesCommandOutputInErrors(t *testing.T) {
	runner := commandRunner{path: "/bin/sh"}
	_, err := runner.Run(context.Background(), "-c", "printf 'useful failure' >&2; exit 1")
	if err == nil || !strings.Contains(err.Error(), "useful failure") {
		t.Fatalf("error = %v, want command output", err)
	}
}

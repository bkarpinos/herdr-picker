package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(context.Context, ...string) ([]byte, error)
}

type commandRunner struct {
	path string
}

func NewRunner(path string) Runner {
	if path == "" {
		path = "herdr"
	}
	return commandRunner{path: path}
}

func (r commandRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	output, err := exec.CommandContext(ctx, r.path, args...).CombinedOutput()
	if err != nil {
		if detail := strings.TrimSpace(string(output)); detail != "" {
			return nil, fmt.Errorf("%w: %s", err, detail)
		}
		return nil, err
	}
	return output, nil
}

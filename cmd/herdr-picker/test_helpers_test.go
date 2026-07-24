package main

import (
	"context"
	"errors"
	"strings"
	"sync"
)

type testRunner struct {
	mu        sync.Mutex
	responses map[string][]byte
	errors    map[string]error
	calls     [][]string
}

func (r *testRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, append([]string(nil), args...))
	key := strings.Join(args, " ")
	if err := r.errors[key]; err != nil {
		return nil, err
	}
	if output, ok := r.responses[key]; ok {
		return output, nil
	}
	return nil, errors.New("unexpected command: " + key)
}

func (r *testRunner) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.calls)
}

func (r *testRunner) lastCall() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.calls) == 0 {
		return ""
	}
	return strings.Join(r.calls[len(r.calls)-1], " ")
}

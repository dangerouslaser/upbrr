// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package dupe

import (
	"testing"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/trackers"
	trackerimpl "github.com/autobrr/upbrr/internal/trackers/impl"
)

func TestBuildHandlersCoversKnownTrackers(t *testing.T) {
	t.Parallel()
	registry, err := trackerimpl.NewRegistry()
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	handlers := NewServiceWithRegistry(config.Config{}, nil, registry).handlers

	for _, tracker := range trackers.KnownTrackers() {
		if tracker == "MANUAL" {
			continue
		}
		if _, ok := handlers[tracker]; !ok {
			t.Fatalf("expected handler for %s", tracker)
		}
	}
}

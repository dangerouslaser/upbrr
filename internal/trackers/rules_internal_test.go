// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package trackers

import (
	"testing"

	"github.com/autobrr/upbrr/pkg/api"
)

func TestResolveCategoryIgnoresEmptyTVMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata api.ExternalMetadata
	}{
		{name: "TVDB", metadata: api.ExternalMetadata{TVDB: &api.TVDBMetadata{}}},
		{name: "TVmaze", metadata: api.ExternalMetadata{TVmaze: &api.TVmazeMetadata{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			meta := api.PreparedMetadata{
				ExternalMetadata: tt.metadata,
				Release:          api.ReleaseInfo{Category: "movie"},
			}
			if got := resolveCategory(meta); got != "movie" {
				t.Fatalf("expected movie fallback, got %q", got)
			}
		})
	}
}

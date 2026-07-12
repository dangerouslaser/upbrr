// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package azfamily

import (
	"context"
	"testing"

	"github.com/autobrr/upbrr/pkg/api"
)

func evaluateSiteRules(t *testing.T, tracker string, meta api.PreparedMetadata) []api.RuleFailure {
	t.Helper()
	return New(tracker).evaluateRules(context.Background(), meta, api.NopLogger{})
}

func TestEvaluateRulesAZRedirectsEnglishTerritories(t *testing.T) {
	t.Parallel()

	failures := evaluateSiteRules(t, "AZ", api.PreparedMetadata{
		ExternalIDs: api.ExternalIDs{Category: "MOVIE"},
		ExternalMetadata: api.ExternalMetadata{
			TMDB: &api.TMDBMetadata{OriginCountry: []string{"US"}},
		},
	})
	if len(failures) == 0 {
		t.Fatal("expected AZ rule failure")
	}
}

func TestEvaluateRulesCZRejectsAsianContent(t *testing.T) {
	t.Parallel()

	failures := evaluateSiteRules(t, "CZ", api.PreparedMetadata{
		ExternalIDs: api.ExternalIDs{Category: "MOVIE"},
		ExternalMetadata: api.ExternalMetadata{
			TMDB: &api.TMDBMetadata{OriginCountry: []string{"JP"}},
		},
	})
	if len(failures) == 0 {
		t.Fatal("expected CZ rule failure")
	}
}

func TestEvaluateRulesPHDRejectsSDAndBlockedGroup(t *testing.T) {
	t.Parallel()

	failures := evaluateSiteRules(t, "PHD", api.PreparedMetadata{
		ExternalIDs: api.ExternalIDs{Category: "MOVIE"},
		Release:     api.ReleaseInfo{Resolution: "480p"},
		Container:   "avi",
		Tag:         "-RARBG",
	})
	if len(failures) < 2 {
		t.Fatalf("expected multiple PHD failures, got %v", failures)
	}
}

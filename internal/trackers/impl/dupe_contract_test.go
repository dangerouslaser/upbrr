// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package impl

import (
	"context"
	"strings"
	"testing"

	"github.com/autobrr/upbrr/internal/config"
	arimpl "github.com/autobrr/upbrr/internal/trackers/impl/ar"
	bhdtvimpl "github.com/autobrr/upbrr/internal/trackers/impl/bhdtv"
	dcimpl "github.com/autobrr/upbrr/internal/trackers/impl/dc"
	gpwimpl "github.com/autobrr/upbrr/internal/trackers/impl/gpw"
	mtvimpl "github.com/autobrr/upbrr/internal/trackers/impl/mtv"
	nblimpl "github.com/autobrr/upbrr/internal/trackers/impl/nbl"
	rtfimpl "github.com/autobrr/upbrr/internal/trackers/impl/rtf"
	spdimpl "github.com/autobrr/upbrr/internal/trackers/impl/spd"
	tlimpl "github.com/autobrr/upbrr/internal/trackers/impl/tl"
	tvcimpl "github.com/autobrr/upbrr/internal/trackers/impl/tvc"
	"github.com/autobrr/upbrr/pkg/api"
)

type dupeContractSearcher interface {
	Search(context.Context, api.PreparedMetadata, string) ([]api.DupeEntry, []string, error)
}

func hasContractSkipReason(notes []string) bool {
	for _, note := range notes {
		trimmed := strings.TrimSpace(note)
		if !strings.HasPrefix(strings.ToLower(trimmed), "skip: ") {
			continue
		}
		reason := strings.TrimSpace(trimmed[len("skip: "):])
		return reason != ""
	}
	return false
}

func TestAPIHandlersMissingCredentialsSkip(t *testing.T) {
	t.Parallel()
	meta := api.PreparedMetadata{SourcePath: "x", ExternalIDs: api.ExternalIDs{TMDBID: 1, IMDBID: 1}}
	cases := []struct {
		name    string
		handler dupeContractSearcher
	}{
		{name: "AR", handler: arimpl.New().NewDupeSearcher(config.Config{}, nil, api.NopLogger{})},
		{name: "DC", handler: dcimpl.New().NewDupeSearcher(config.Config{}, nil, api.NopLogger{})},
		{name: "GPW", handler: gpwimpl.New().NewDupeSearcher(config.Config{}, nil, api.NopLogger{})},
		{name: "NBL", handler: nblimpl.New().NewDupeSearcher(config.Config{}, nil, api.NopLogger{})},
		{name: "MTV", handler: mtvimpl.New().NewDupeSearcher(config.Config{}, nil, api.NopLogger{})},
		{name: "RTF", handler: rtfimpl.New().NewDupeSearcher(config.Config{}, nil, api.NopLogger{})},
		{name: "SPD", handler: spdimpl.New().NewDupeSearcher(config.Config{}, nil, api.NopLogger{})},
		{name: "TL", handler: tlimpl.New().NewDupeSearcher(config.Config{}, nil, api.NopLogger{})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, notes, err := tc.handler.Search(context.Background(), meta, tc.name)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !hasContractSkipReason(notes) {
				t.Fatalf("expected skip reason notes, got %#v", notes)
			}
		})
	}
}

func TestBHDTVHandlerReturnsManualMessage(t *testing.T) {
	t.Parallel()
	entries, notes, err := bhdtvimpl.New().Search(context.Background(), api.PreparedMetadata{}, "BHDTV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one manual entry")
	}
	if len(notes) == 0 {
		t.Fatalf("expected notes")
	}
}

func TestTVCHandlerSkipsRestrictedContent(t *testing.T) {
	t.Parallel()
	meta := api.PreparedMetadata{Release: api.ReleaseInfo{Resolution: "2160p"}}
	definition := tvcimpl.New()
	searcher, ok := definition.(interface {
		Search(context.Context, api.PreparedMetadata, string) ([]api.DupeEntry, []string, error)
	})
	if !ok {
		t.Fatal("expected TVC dupe search capability")
	}
	_, notes, err := searcher.Search(context.Background(), meta, "TVC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasContractSkipReason(notes) {
		t.Fatalf("expected skip reason")
	}
}

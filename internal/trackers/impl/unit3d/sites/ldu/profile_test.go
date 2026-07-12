package ldu

import (
	"strings"
	"testing"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/pkg/api"
)

func TestBuildNameUsesFirstParseableLanguages(t *testing.T) {
	meta := api.PreparedMetadata{ReleaseName: "Example.Release.2026.1080p.WEB-DL.DD5.1.H264-GRP", ExternalIDs: api.ExternalIDs{Category: "MOVIE"}, AudioLanguages: []string{"", "Japanese", "English"}, SubtitleLanguages: []string{"", "English"}, ExternalMetadata: api.ExternalMetadata{TMDB: &api.TMDBMetadata{OriginalLanguage: "ja"}}}
	got := Profile().Site.BuildName(meta, config.TrackerConfig{})
	if !strings.Contains(got, "[JPN]") || !strings.Contains(got, "[Subs ENG]") {
		t.Fatalf("name = %q", got)
	}
}

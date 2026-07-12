// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package spd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/pkg/api"
)

type dupeSearcher struct {
	cfg  config.Config
	http *http.Client
}

func (Definition) NewDupeSearcher(cfg config.Config, httpClient *http.Client, _ api.Logger) trackers.DupeSearcher {
	return &dupeSearcher{cfg: cfg, http: httpClient}
}

func (s *dupeSearcher) Search(ctx context.Context, meta api.PreparedMetadata, _ string) ([]api.DupeEntry, []string, error) {
	apiKey := spdAPIKey(s.cfg)
	if apiKey == "" {
		return nil, spdSkip("missing api_key for tracker"), nil
	}
	params := url.Values{}
	switch {
	case meta.ExternalIDs.IMDBID != 0:
		params.Set("imdbId", strconv.Itoa(meta.ExternalIDs.IMDBID))
	case strings.TrimSpace(meta.Release.Title) != "":
		params.Set("search", strings.TrimSpace(meta.Release.Title))
	default:
		return nil, spdSkip("missing imdb/title for SPD dupe search"), nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://speedapp.io/api/torrent", nil)
	if err != nil {
		return nil, spdSkip("SPD search failed"), nil
	}
	req.URL.RawQuery = params.Encode()
	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Accept", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, spdSkip("SPD search failed"), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, spdSkip("SPD search failed"), nil
	}
	var items []map[string]any
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&items); err != nil {
		return nil, spdSkip("SPD search failed"), nil
	}
	entries := make([]api.DupeEntry, 0, len(items))
	for _, item := range items {
		id := spdString(item["id"])
		entry := api.DupeEntry{Name: spdString(item["name"]), ID: id, Link: "https://speedapp.io/browse/" + id + "/"}
		if size := spdInt64(item["size"]); size > 0 {
			entry.SizeKnown, entry.SizeBytes = true, size
		}
		entries = append(entries, entry)
	}
	return entries, nil, nil
}

func spdAPIKey(cfg config.Config) string {
	for name, entry := range cfg.Trackers.Trackers {
		if strings.EqualFold(strings.TrimSpace(name), "SPD") {
			return strings.TrimSpace(entry.APIKey)
		}
	}
	return ""
}
func spdString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}
func spdInt64(value any) int64 {
	switch typed := value.(type) {
	case json.Number:
		parsed, _ := typed.Int64()
		return parsed
	case float64:
		return int64(typed)
	default:
		return 0
	}
}
func spdSkip(reason string) []string { return []string{"skip: " + reason} }

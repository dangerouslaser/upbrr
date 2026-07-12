// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package dc

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
	cfg      config.Config
	http     *http.Client
	endpoint string
}

func (Definition) NewDupeSearcher(cfg config.Config, httpClient *http.Client, _ api.Logger) trackers.DupeSearcher {
	return &dupeSearcher{cfg: cfg, http: httpClient, endpoint: "https://digitalcore.club/api/v1/torrents"}
}

func (s *dupeSearcher) Search(ctx context.Context, meta api.PreparedMetadata, _ string) ([]api.DupeEntry, []string, error) {
	apiKey := dcAPIKey(s.cfg)
	if apiKey == "" {
		return nil, dcSkip("missing api_key for tracker"), nil
	}
	if meta.ExternalIDs.IMDBID == 0 {
		return nil, dcSkip("missing imdb id for DC dupe search"), nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.endpoint, nil)
	if err != nil {
		return nil, dcSkip("DC search failed"), nil
	}
	req.URL.RawQuery = url.Values{"searchText": {"tt" + strconv.Itoa(meta.ExternalIDs.IMDBID)}}.Encode()
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, dcSkip("DC search failed"), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, dcSkip("DC search failed"), nil
	}
	var items []map[string]any
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&items); err != nil {
		return nil, dcSkip("DC search failed"), nil
	}
	entries := make([]api.DupeEntry, 0, len(items))
	for _, item := range items {
		id := dcString(item["id"])
		entry := api.DupeEntry{Name: dcString(item["name"]), ID: id, Link: "https://digitalcore.club/torrent/" + id + "/"}
		if size := dcInt64(item["size"]); size > 0 {
			entry.SizeKnown, entry.SizeBytes = true, size
		}
		entries = append(entries, entry)
	}
	return entries, nil, nil
}

func dcAPIKey(cfg config.Config) string {
	for name, entry := range cfg.Trackers.Trackers {
		if strings.EqualFold(strings.TrimSpace(name), "DC") {
			return strings.TrimSpace(entry.APIKey)
		}
	}
	return ""
}

func dcString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func dcInt64(value any) int64 {
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

func dcSkip(reason string) []string { return []string{"skip: " + reason} }

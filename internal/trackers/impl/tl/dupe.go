// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package tl

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
	if !tlConfigured(s.cfg) {
		return nil, tlSkip("missing passkey for tracker"), nil
	}
	query := strings.TrimSpace(meta.Release.Title)
	if query == "" {
		query = strings.TrimSpace(meta.ReleaseName)
	}
	if query == "" {
		return nil, tlSkip("missing title for TL dupe search"), nil
	}
	endpoint := "https://www.torrentleech.org/torrents/browse/list/query/" + url.PathEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, tlSkip("TL search failed"), nil
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, tlSkip("TL search failed"), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, tlSkip("TL search failed"), nil
	}
	var payload struct {
		Torrents []map[string]any `json:"torrentList"`
	}
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, tlSkip("TL search failed"), nil
	}
	entries := make([]api.DupeEntry, 0, len(payload.Torrents))
	for _, item := range payload.Torrents {
		id := tlString(item["fid"])
		entry := api.DupeEntry{Name: tlString(item["name"]), ID: id, Link: "https://www.torrentleech.org/torrent/" + id}
		if size := tlInt64(item["size"]); size > 0 {
			entry.SizeKnown, entry.SizeBytes = true, size
		}
		entries = append(entries, entry)
	}
	return entries, nil, nil
}

func tlConfigured(cfg config.Config) bool {
	for name, entry := range cfg.Trackers.Trackers {
		if strings.EqualFold(strings.TrimSpace(name), "TL") {
			return strings.TrimSpace(entry.Passkey) != ""
		}
	}
	return false
}

func tlString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func tlInt64(value any) int64 {
	switch typed := value.(type) {
	case json.Number:
		parsed, _ := typed.Int64()
		return parsed
	case float64:
		return int64(typed)
	case string:
		parsed, _ := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		return parsed
	default:
		return 0
	}
}

func tlSkip(reason string) []string { return []string{"skip: " + reason} }

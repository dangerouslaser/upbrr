// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package gpw

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
	return &dupeSearcher{cfg: cfg, http: httpClient, endpoint: "https://greatposterwall.com/api.php"}
}

func (s *dupeSearcher) Search(ctx context.Context, meta api.PreparedMetadata, _ string) ([]api.DupeEntry, []string, error) {
	apiKey := gpwAPIKey(s.cfg)
	if apiKey == "" {
		return nil, gpwSkip("missing api_key for tracker"), nil
	}
	if meta.ExternalIDs.IMDBID == 0 {
		return nil, gpwSkip("missing imdb id for GPW dupe search"), nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.endpoint, nil)
	if err != nil {
		return nil, gpwSkip("GPW search failed"), nil
	}
	req.URL.RawQuery = url.Values{"api_key": {apiKey}, "action": {"torrent"}, "imdbID": {"tt" + strconv.Itoa(meta.ExternalIDs.IMDBID)}}.Encode()
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, gpwSkip("GPW search failed"), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, gpwSkip("GPW search failed"), nil
	}
	var payload struct {
		Status   int              `json:"status"`
		Response []map[string]any `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, gpwSkip("GPW search failed"), nil
	}
	if payload.Status != http.StatusOK {
		return nil, nil, nil
	}
	entries := make([]api.DupeEntry, 0, len(payload.Response))
	for _, item := range payload.Response {
		parts := []string{gpwString(item["Name"]), gpwString(item["Year"]), gpwString(item["Resolution"]), gpwString(item["Source"]), gpwString(item["Processing"]), gpwString(item["RemasterTitle"]), gpwString(item["Codec"])}
		entries = append(entries, api.DupeEntry{Name: strings.Join(strings.Fields(strings.Join(parts, " ")), " ")})
	}
	return entries, nil, nil
}

func gpwAPIKey(cfg config.Config) string {
	for name, entry := range cfg.Trackers.Trackers {
		if strings.EqualFold(strings.TrimSpace(name), "GPW") {
			return strings.TrimSpace(entry.APIKey)
		}
	}
	return ""
}

func gpwString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func gpwSkip(reason string) []string { return []string{"skip: " + reason} }

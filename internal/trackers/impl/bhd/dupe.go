// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package bhd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/pkg/api"
)

var seasonPattern = regexp.MustCompile(`(?i)S(\d{1,2})`)

type dupeSearcher struct {
	cfg     config.Config
	http    *http.Client
	baseURL string
}

func (d *Definition) NewDupeSearcher(cfg config.Config, httpClient *http.Client, _ api.Logger) trackers.DupeSearcher {
	return &dupeSearcher{cfg: cfg, http: httpClient, baseURL: "https://beyond-hd.me/api/torrents/"}
}

func (s *dupeSearcher) Search(ctx context.Context, meta api.PreparedMetadata, _ string) ([]api.DupeEntry, []string, error) {
	cfg, apiKey := bhdConfig(s.cfg)
	if apiKey == "" {
		return nil, bhdSkip("missing api_key for tracker"), nil
	}
	tmdbID, imdbID := meta.ExternalIDs.TMDBID, bhdIMDB(meta.ExternalIDs.IMDBID)
	if tmdbID == 0 && imdbID == "" {
		return nil, bhdSkip("missing tmdb/imdb id for BHD dupe search"), nil
	}
	category, tmdbPrefix := "Movies", "movie"
	if strings.EqualFold(meta.ExternalIDs.Category, "TV") {
		category, tmdbPrefix = "TV", "tv"
	}
	payload := map[string]any{"action": "search", "categories": category}
	if searchType, ok := SearchType(meta); ok {
		payload["types"] = searchType
	} else {
		payload["types"] = nil
	}
	if IsSD(meta) {
		payload["categories"], payload["types"] = nil, nil
	}
	if tmdbID != 0 {
		payload["tmdb_id"] = tmdbPrefix + "/" + strconv.Itoa(tmdbID)
	} else {
		payload["imdb_id"] = imdbID
	}
	if season := bhdSeason(meta); season != "" && category == "TV" {
		payload["search"] = season
	}
	if rss := strings.TrimSpace(cfg.BhdRSSKey); rss != "" {
		payload["rsskey"] = rss
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, bhdSkip("BHD request failed"), nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+apiKey, bytes.NewReader(body))
	if err != nil {
		return nil, bhdSkip("BHD request failed"), nil
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, bhdSkip("BHD request failed"), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, bhdSkip("BHD search failed"), nil
	}
	var decoded map[string]any
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil || len(decoded) == 0 {
		return nil, bhdSkip("BHD search failed"), nil
	}
	if bhdInt(decoded["status_code"]) == 0 {
		return nil, bhdSkip("BHD api rejected search"), nil
	}
	return bhdEntries(decoded), nil, nil
}

func bhdEntries(payload map[string]any) []api.DupeEntry {
	results, _ := payload["results"].([]any)
	entries := make([]api.DupeEntry, 0, len(results))
	for _, raw := range results {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		entry := api.DupeEntry{Name: bhdString(item["name"]), Link: bhdString(item["url"])}
		if size := bhdInt(item["size"]); size > 0 {
			entry.SizeKnown, entry.SizeBytes = true, size
		}
		if bhdInt(item["dv"]) == 1 {
			entry.Flags = append(entry.Flags, "DV")
		}
		if bhdInt(item["hdr10"]) == 1 || bhdInt(item["hdr10+"]) == 1 {
			entry.Flags = append(entry.Flags, "HDR")
		}
		entries = append(entries, entry)
	}
	return entries
}

func bhdConfig(cfg config.Config) (config.TrackerConfig, string) {
	for name, entry := range cfg.Trackers.Trackers {
		if strings.EqualFold(strings.TrimSpace(name), "BHD") {
			return entry, strings.TrimSpace(entry.APIKey)
		}
	}
	return config.TrackerConfig{}, ""
}

func bhdSeason(meta api.PreparedMetadata) string {
	if meta.ReleaseNameOverrides.Season != nil {
		return normalizeBHDSeason(*meta.ReleaseNameOverrides.Season)
	}
	match := seasonPattern.FindStringSubmatch(meta.ReleaseName)
	if len(match) == 2 {
		return normalizeBHDSeason(match[1])
	}
	return ""
}

func normalizeBHDSeason(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToUpper(trimmed), "S") {
		return strings.ToUpper(trimmed)
	}
	if number, err := strconv.Atoi(trimmed); err == nil {
		return "S" + strconv.Itoa(number)
	}
	return strings.ToUpper(trimmed)
}

func bhdIMDB(id int) string {
	if id == 0 {
		return ""
	}
	return fmt.Sprintf("tt%07d", id)
}

func bhdSkip(reason string) []string { return []string{"skip: " + strings.TrimSpace(reason)} }

func bhdString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func bhdInt(value any) int64 {
	switch typed := value.(type) {
	case json.Number:
		parsed, _ := typed.Int64()
		return parsed
	case float64:
		return int64(typed)
	case int:
		return int64(typed)
	case int64:
		return typed
	case string:
		parsed, _ := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		return parsed
	default:
		return 0
	}
}

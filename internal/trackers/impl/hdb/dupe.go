// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package hdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/redaction"
	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/pkg/api"
)

type dupeSearcher struct {
	cfg      config.Config
	http     *http.Client
	logger   api.Logger
	endpoint string
}

func (d *Definition) NewDupeSearcher(cfg config.Config, httpClient *http.Client, logger api.Logger) trackers.DupeSearcher {
	if logger == nil {
		logger = api.NopLogger{}
	}
	return &dupeSearcher{
		cfg:      cfg,
		http:     httpClient,
		logger:   logger,
		endpoint: "https://hdbits.org/api/torrents",
	}
}

func (s *dupeSearcher) Search(ctx context.Context, meta api.PreparedMetadata, _ string) ([]api.DupeEntry, []string, error) {
	if s.http == nil {
		return nil, hdbSkip("HDB handler misconfigured: no HTTP client"), nil
	}
	username, passkey := hdbCredentials(s.cfg)
	if username == "" || passkey == "" {
		return nil, hdbSkip("missing username/passkey for tracker"), nil
	}
	payload := map[string]any{
		"username": username,
		"passkey":  passkey,
		"category": hdbCategoryID(meta),
		"codec":    hdbCodecID(meta),
		"medium":   hdbMediumID(meta),
	}
	searchMethod := "id"
	if meta.ExternalIDs.IMDBID != 0 {
		payload["imdb"] = map[string]any{"id": fmt.Sprintf("%07d", meta.ExternalIDs.IMDBID)}
	} else if isHDBTVCategory(meta) && meta.ExternalIDs.TVDBID != 0 {
		payload["tvdb"] = map[string]any{"id": meta.ExternalIDs.TVDBID}
	}
	if _, hasIMDB := payload["imdb"]; !hasIMDB {
		if _, hasTVDB := payload["tvdb"]; !hasTVDB {
			query := firstHDBText(meta.ReleaseName, meta.Filename, meta.Release.Title)
			if query == "" {
				s.logger.Warnf("dupechecking: HDB missing imdb/tvdb IDs and search text for %s", meta.SourcePath)
				return nil, hdbSkip("missing imdb/tvdb id for HDB dupe search"), nil
			}
			payload["search"], searchMethod = query, "text_fallback"
			s.logger.Debugf("dupechecking: HDB falling back to text search for %s", meta.SourcePath)
		}
	}
	if logPayload, err := json.Marshal(redaction.RedactPrivateInfo(payload, nil)); err != nil {
		s.logger.Debugf("dupechecking: HDB search payload_marshal_failed=%v source=%s", err, meta.SourcePath)
	} else {
		s.logger.Debugf("dupechecking: HDB search payload=%s source=%s", string(logPayload), meta.SourcePath)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, hdbSkip("HDB request failed"), nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, hdbSkip("HDB request failed"), nil
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		s.logger.Warnf("dupechecking: HDB request failed for %s: %v", meta.SourcePath, err)
		return nil, hdbSkip("HDB request failed"), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		s.logger.Warnf("dupechecking: HDB search failed for %s with status=%d", meta.SourcePath, resp.StatusCode)
		return nil, hdbSkip("HDB search failed"), nil
	}
	var body map[string]any
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&body); err != nil || len(body) == 0 {
		return nil, hdbSkip("HDB search failed"), nil
	}
	if hdbInt(body["status"]) != 0 {
		s.logger.Warnf("dupechecking: HDB API rejected search for %s", meta.SourcePath)
		return nil, hdbSkip("HDB api rejected search"), nil
	}
	items, _ := body["data"].([]any)
	entries := make([]api.DupeEntry, 0, len(items))
	for _, rawItem := range items {
		item, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}
		id, filename := hdbString(item["id"]), hdbString(item["filename"])
		entry := api.DupeEntry{
			Name:      hdbString(item["name"]),
			ID:        id,
			Link:      "https://hdbits.org/details.php?id=" + id,
			Download:  "https://hdbits.org/download.php/" + url.QueryEscape(filename) + "?id=" + id + "&passkey=" + passkey,
			FileCount: hdbInt(item["numfiles"]),
		}
		if size := hdbInt(item["size"]); size > 0 {
			entry.SizeKnown, entry.SizeBytes = true, int64(size)
		}
		entries = append(entries, entry)
	}
	s.logger.Debugf("dupechecking: HDB returned %d entries for %s method=%s", len(entries), meta.SourcePath, searchMethod)
	return entries, nil, nil
}

func firstHDBText(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func hdbSkip(reason string) []string { return []string{"skip: " + strings.TrimSpace(reason)} }

// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package btn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

func (d *Definition) NewDupeSearcher(cfg config.Config, httpClient *http.Client, _ api.Logger) trackers.DupeSearcher {
	return &dupeSearcher{cfg: cfg, http: httpClient, endpoint: "https://api.broadcasthe.net/"}
}

func (s *dupeSearcher) Search(ctx context.Context, meta api.PreparedMetadata, _ string) ([]api.DupeEntry, []string, error) {
	token := strings.TrimSpace(config.ResolveBTNAPIToken(s.cfg))
	if token == "" {
		return nil, btnSkip("missing api_key for tracker"), nil
	}
	if !isTV(meta) {
		return nil, btnSkip("BTN only supports TV dupe search"), nil
	}
	filter := make(map[string]any)
	switch {
	case trackerID(meta) != "":
		filter["id"] = trackerID(meta)
	case meta.ExternalIDs.IMDBID != 0:
		filter["imdb"] = fmt.Sprintf("tt%07d", meta.ExternalIDs.IMDBID)
	case meta.ExternalIDs.TVDBID != 0:
		filter["tvdb"] = meta.ExternalIDs.TVDBID
	case searchTitle(meta) != "":
		filter["searchstr"] = searchTitle(meta)
	default:
		return nil, btnSkip("missing btn/imdb/tvdb id and title for BTN dupe search"), nil
	}
	payload := map[string]any{"jsonrpc": "2.0", "id": "upbrr-btn-search", "method": "getTorrentsSearch", "params": []any{token, filter, 50}}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, btnSkip("BTN request failed"), nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, btnSkip("BTN request failed"), nil
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, btnSkip("BTN request failed"), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, btnSkip("BTN search failed"), nil
	}
	var response map[string]any
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		return nil, btnSkip("BTN search failed"), nil
	}
	if errorPayload, ok := response["error"].(map[string]any); ok && len(errorPayload) > 0 {
		return nil, nil, nil
	}
	result, _ := response["result"].(map[string]any)
	torrents, _ := result["torrents"].(map[string]any)
	entries := make([]api.DupeEntry, 0, len(torrents))
	for id, rawTorrent := range torrents {
		torrent, ok := rawTorrent.(map[string]any)
		if !ok {
			continue
		}
		entry := api.DupeEntry{Name: releaseName(id, torrent), ID: strings.TrimSpace(id), Link: torrentLink(id, torrent), Res: btnString(first(torrent, "Resolution", "resolution")), Type: btnString(first(torrent, "Source", "source", "Type", "type"))}
		if size := btnInt(first(torrent, "Size", "size")); size > 0 {
			entry.SizeKnown, entry.SizeBytes = true, size
		}
		entry.Flags = flags(torrent)
		entries = append(entries, entry)
	}
	return entries, nil, nil
}

func isTV(meta api.PreparedMetadata) bool {
	category := strings.ToUpper(strings.TrimSpace(meta.ExternalIDs.Category))
	if category == "" {
		category = strings.ToUpper(strings.TrimSpace(meta.MediaInfoCategory))
	}
	return category == "TV"
}

func trackerID(meta api.PreparedMetadata) string {
	for key, value := range meta.TrackerIDs {
		if strings.EqualFold(strings.TrimSpace(key), "BTN") {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func searchTitle(meta api.PreparedMetadata) string {
	candidates := []string{strings.TrimSpace(meta.Release.Title)}
	if meta.ExternalMetadata.TVDB != nil {
		candidates = append(candidates, strings.TrimSpace(meta.ExternalMetadata.TVDB.Name), strings.TrimSpace(meta.ExternalMetadata.TVDB.NameEnglish))
	}
	if meta.ExternalMetadata.TVmaze != nil {
		candidates = append(candidates, strings.TrimSpace(meta.ExternalMetadata.TVmaze.Name))
	}
	candidates = append(candidates, strings.TrimSpace(meta.Filename), strings.TrimSpace(meta.ReleaseName))
	for _, candidate := range candidates {
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

func releaseName(id string, torrent map[string]any) string {
	for _, candidate := range []string{btnString(first(torrent, "ReleaseName", "releaseName")), btnString(first(torrent, "SceneName", "Name", "name")), btnString(first(torrent, "Series", "series")), strings.TrimSpace(id)} {
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

func torrentLink(id string, torrent map[string]any) string {
	groupID := btnString(first(torrent, "GroupID", "groupId"))
	if groupID == "" || strings.TrimSpace(id) == "" {
		return ""
	}
	return "https://broadcasthe.net/torrents.php?id=" + groupID + "&torrentid=" + strings.TrimSpace(id)
}

func flags(torrent map[string]any) []string {
	out := make([]string, 0, 2)
	for _, value := range []string{btnString(first(torrent, "HDR", "hdr")), btnString(first(torrent, "DolbyVision", "dolbyVision", "DV", "dv"))} {
		upper := strings.ToUpper(strings.TrimSpace(value))
		switch upper {
		case "", "0", "FALSE", "NO", "1", "TRUE", "YES":
			continue
		}
		out = append(out, upper)
	}
	return out
}

func first(values map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value
		}
	}
	return nil
}

func btnString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func btnSkip(reason string) []string { return []string{"skip: " + strings.TrimSpace(reason)} }

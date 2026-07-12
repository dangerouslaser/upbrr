// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package pts

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	xhtml "golang.org/x/net/html"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/cookies"
	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/internal/trackers/impl/commonhttp"
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
	if meta.ExternalIDs.IMDBID == 0 {
		return nil, ptsSkip("missing IMDb ID for PTS dupe search"), nil
	}
	baseURL := ptsBaseURL(s.cfg)
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, ptsSkip("PTS search failed"), nil
	}
	trackerCookies, err := cookies.LoadTrackerHTTPCookies(ctx, s.cfg.MainSettings.DBPath, "PTS", parsed.Hostname())
	if err != nil {
		return nil, ptsSkip("missing valid PTS cookies"), nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/torrents.php", nil)
	if err != nil {
		return nil, ptsSkip("PTS search failed"), nil
	}
	req.URL.RawQuery = url.Values{"incldead": {"1"}, "search": {fmt.Sprintf("tt%07d", meta.ExternalIDs.IMDBID)}, "search_area": {"4"}}.Encode()
	req.Header.Set("User-Agent", "upbrr")
	commonhttp.ApplyCookies(req, trackerCookies)
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, ptsSkip("PTS search failed"), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, ptsSkip("PTS search failed"), nil
	}
	root, err := xhtml.Parse(resp.Body)
	if err != nil {
		return nil, ptsSkip("PTS search failed"), nil
	}
	entries := make([]api.DupeEntry, 0)
	ptsWalk(root, func(node *xhtml.Node) {
		if node.Type != xhtml.ElementNode || node.Data != "b" {
			return
		}
		if name := strings.TrimSpace(ptsNodeText(node)); name != "" {
			entries = append(entries, api.DupeEntry{Name: name})
		}
	})
	return entries, nil, nil
}

func ptsBaseURL(cfg config.Config) string {
	for name, entry := range cfg.Trackers.Trackers {
		if strings.EqualFold(strings.TrimSpace(name), "PTS") && strings.TrimSpace(entry.URL) != "" {
			return strings.TrimRight(strings.TrimSpace(entry.URL), "/")
		}
	}
	return "https://www.ptskit.org"
}

func ptsWalk(node *xhtml.Node, visit func(*xhtml.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		ptsWalk(child, visit)
	}
}

func ptsNodeText(node *xhtml.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == xhtml.TextNode {
		return node.Data
	}
	var builder strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		builder.WriteString(ptsNodeText(child))
	}
	return builder.String()
}

func ptsSkip(reason string) []string { return []string{"skip: " + reason} }

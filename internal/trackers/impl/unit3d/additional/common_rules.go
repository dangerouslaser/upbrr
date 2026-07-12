// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package additional

import (
	"strings"

	"github.com/autobrr/upbrr/pkg/api"
)

var resolutionOrder = map[string]int{
	"480i":  1,
	"480p":  2,
	"576i":  3,
	"576p":  4,
	"720p":  5,
	"1080i": 6,
	"1080p": 7,
	"1440p": 8,
	"2160p": 9,
	"4320p": 10,
	"8640p": 11,
}

func isDolbyVisionOnly(meta api.PreparedMetadata) bool {
	if meta.WebDV {
		return true
	}
	hdr := strings.ToUpper(strings.TrimSpace(meta.HDR))
	return strings.Contains(hdr, "DV") && !strings.Contains(hdr, "HDR")
}

func resolveResolution(meta api.PreparedMetadata) string {
	resolution := strings.TrimSpace(meta.Release.Resolution)
	if resolution == "" {
		resolution = detectResolution(meta.ReleaseName)
	}
	return resolution
}

func detectResolution(value string) string {
	clean := strings.ToLower(value)
	for _, candidate := range []string{"8640p", "4320p", "2160p", "1440p", "1080p", "1080i", "720p", "576p", "576i", "480p", "480i"} {
		if strings.Contains(clean, candidate) {
			return candidate
		}
	}
	return ""
}

func resolveType(meta api.PreparedMetadata) string {
	typeValue := strings.ToUpper(strings.TrimSpace(meta.Type))
	if typeValue == "" {
		typeValue = strings.ToUpper(strings.TrimSpace(meta.Release.Type))
	}
	return typeValue
}

func resolveGroup(meta api.PreparedMetadata) string {
	if group := strings.TrimSpace(meta.Release.Group); group != "" {
		return group
	}
	return strings.TrimPrefix(strings.TrimSpace(meta.Tag), "-")
}

func isDiscType(value string) bool {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "BDMV", "DVD", "HDDVD":
		return true
	default:
		return false
	}
}

func isAdultContent(meta api.PreparedMetadata) bool {
	candidates := append([]string{}, collectGenres(meta)...)
	candidates = append(candidates, collectKeywords(meta)...)
	for _, token := range candidates {
		switch strings.ToLower(strings.TrimSpace(token)) {
		case "adult", "porn", "pornography", "xxx", "erotic":
			return true
		}
	}
	return false
}

func isAnime(meta api.PreparedMetadata) bool {
	if meta.ExternalMetadata.TMDB != nil && externalMetadataMatchesCurrentSource(meta) && meta.ExternalMetadata.TMDB.Anime {
		return true
	}
	return containsAny(collectKeywords(meta), []string{"anime"})
}

func isAnimation(meta api.PreparedMetadata) bool {
	return containsAny(collectGenres(meta), []string{"animation"})
}

func collectGenres(meta api.PreparedMetadata) []string {
	values := []string{}
	values = append(values, splitList(meta.Release.Genre)...)
	if meta.ExternalMetadata.TMDB != nil && externalMetadataMatchesCurrentSource(meta) {
		values = append(values, splitList(meta.ExternalMetadata.TMDB.Genres)...)
	}
	if meta.ExternalMetadata.IMDB != nil && externalMetadataMatchesCurrentSource(meta) {
		values = append(values, splitList(meta.ExternalMetadata.IMDB.Genres)...)
	}
	return normalizeStrings(values)
}

func collectKeywords(meta api.PreparedMetadata) []string {
	values := []string{}
	if meta.ExternalMetadata.TMDB != nil && externalMetadataMatchesCurrentSource(meta) {
		values = append(values, splitList(meta.ExternalMetadata.TMDB.Keywords)...)
	}
	return normalizeStrings(values)
}

func externalMetadataMatchesCurrentSource(meta api.PreparedMetadata) bool {
	storedSource := strings.TrimSpace(meta.ExternalMetadata.SourcePath)
	return storedSource == "" || strings.EqualFold(storedSource, strings.TrimSpace(meta.SourcePath))
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	return out
}

func containsAny(values []string, targets []string) bool {
	if len(values) == 0 || len(targets) == 0 {
		return false
	}
	targetSet := make(map[string]bool, len(targets))
	for _, target := range targets {
		targetSet[strings.ToLower(strings.TrimSpace(target))] = true
	}
	for _, value := range values {
		if targetSet[strings.ToLower(strings.TrimSpace(value))] {
			return true
		}
	}
	return false
}

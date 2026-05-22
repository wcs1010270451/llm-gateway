package service

import (
	"net/url"
	"strings"
)

func buildUpstreamURL(baseURL string, endpointPath string) string {
	if strings.TrimSpace(baseURL) == "" {
		return "/" + strings.TrimLeft(strings.TrimSpace(endpointPath), "/")
	}

	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(endpointPath, "/")
	}
	baseSegments := splitPathSegments(parsed.Path)
	endpointSegments := splitPathSegments(endpointPath)
	if len(endpointSegments) == 0 {
		return parsed.String()
	}
	if pathHasSuffix(baseSegments, endpointSegments) {
		return parsed.String()
	}
	if len(baseSegments) > 0 && len(endpointSegments) > 0 && baseSegments[len(baseSegments)-1] == endpointSegments[0] {
		endpointSegments = endpointSegments[1:]
	}
	segments := append(append([]string{}, baseSegments...), endpointSegments...)
	parsed.Path = "/" + strings.Join(segments, "/")
	return parsed.String()
}

func splitPathSegments(value string) []string {
	trimmed := strings.Trim(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func pathHasSuffix(baseSegments []string, suffixSegments []string) bool {
	if len(suffixSegments) == 0 || len(baseSegments) < len(suffixSegments) {
		return false
	}
	start := len(baseSegments) - len(suffixSegments)
	for index := range suffixSegments {
		if baseSegments[start+index] != suffixSegments[index] {
			return false
		}
	}
	return true
}

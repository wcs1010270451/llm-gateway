package service

import (
	"log"

	"llm-gateway/backend/internal/repository"
)

func logMissingUsage(route repository.ResolvedRoute, publicModel string, stream bool, requestType string, statusCode int, promptTokens int, completionTokens int, totalTokens int) {
	if statusCode < 200 || statusCode >= 300 || totalTokens > 0 || promptTokens > 0 || completionTokens > 0 {
		return
	}

	log.Printf(
		"[usage] missing token usage provider_id=%d provider=%q provider_slug=%q adapter=%s provider_model_id=%d upstream_model=%q public_model=%q request_type=%s stream=%t status=%d",
		route.Provider.ID,
		route.Provider.Name,
		route.Provider.Slug,
		route.Provider.AdapterType,
		route.ProviderModel.ID,
		route.ProviderModel.UpstreamModel,
		publicModel,
		requestType,
		stream,
		statusCode,
	)
}

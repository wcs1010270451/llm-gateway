package httpapi

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"llm-gateway/backend/internal/api/admin"
	"llm-gateway/backend/internal/api/anthropic"
	authapi "llm-gateway/backend/internal/api/auth"
	"llm-gateway/backend/internal/api/gemini"
	"llm-gateway/backend/internal/api/me"
	"llm-gateway/backend/internal/api/openai"
	"llm-gateway/backend/internal/api/system"
	"llm-gateway/backend/internal/config"
	"llm-gateway/backend/internal/repository"
	"llm-gateway/backend/internal/service"
)

func NewRouter(cfg config.Config, db *gorm.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(corsMiddleware(cfg.FrontendOrigin))
	_ = router.SetTrustedProxies(cfg.GinTrustedProxies)

	providerRepo := repository.NewProviderRepository(db)
	modelRepo := repository.NewModelRepository(db)
	modelFamilyRepo := repository.NewModelFamilyRepository(db)
	userRepo := repository.NewUserRepository(db)
	routingRepo := repository.NewRoutingRepository(db)
	requestLogRepo := repository.NewRequestLogRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	providerKeyCipher := service.NewProviderKeyCipher(cfg.ProviderKeySecret)
	requestLogBodyStore := service.NewRequestLogBodyStore(cfg.LogBodyDir, cfg.LogBodyRetentionDays)
	requestLogBodyStore.StartCleanupLoop(context.Background(), requestLogRepo)
	providerService := service.NewProviderService(providerRepo, providerKeyCipher)
	modelFamilyService := service.NewModelFamilyService(modelFamilyRepo)
	modelService := service.NewModelService(modelRepo, modelFamilyRepo)
	userService := service.NewUserService(userRepo)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, userRepo, providerKeyCipher)
	requestLogService := service.NewRequestLogService(requestLogRepo, requestLogBodyStore)
	adminStatsService := service.NewAdminStatsService(providerRepo, modelRepo, requestLogRepo, userRepo, apiKeyRepo)
	authService := service.NewAuthService(userRepo, cfg.AuthTokenSecret)
	claudeProxyMonitorService := service.NewClaudeProxyMonitorService(providerService)
	anthropicProxyService := service.NewAnthropicProxyService(routingRepo, requestLogRepo, providerKeyCipher, requestLogBodyStore)
	openAIProxyService := service.NewOpenAIProxyService(routingRepo, requestLogRepo, modelRepo, providerKeyCipher, requestLogBodyStore)
	geminiProxyService := service.NewGeminiProxyService(routingRepo, requestLogRepo, providerKeyCipher, requestLogBodyStore)
	userConsoleService := service.NewUserConsoleService(apiKeyRepo, modelRepo, requestLogRepo, requestLogBodyStore, anthropicProxyService, openAIProxyService, geminiProxyService)
	systemHandler := system.NewHandler(db)
	adminHandler := admin.NewHandler(providerService, modelService, modelFamilyService, userService, requestLogService, adminStatsService, claudeProxyMonitorService)
	authHandler := authapi.NewHandler(authService)
	meHandler := me.NewHandler(apiKeyService, userConsoleService)
	anthropicHandler := anthropic.NewHandler(anthropicProxyService, apiKeyService)
	openAIHandler := openai.NewHandler(openAIProxyService, apiKeyService)
	geminiHandler := gemini.NewHandler(geminiProxyService, apiKeyService)

	router.GET("/health", systemHandler.Health)
	router.GET("/ready", systemHandler.Ready)
	router.GET("/v1/models", openAIHandler.Models)
	router.POST("/v1/chat/completions", openAIHandler.ChatCompletions)
	router.POST("/v1/messages", anthropicHandler.Messages)
	router.POST("/v1/gemini/generate-content", geminiHandler.GenerateContent)
	router.POST("/v1/gemini/stream-generate-content", geminiHandler.StreamGenerateContent)
	router.POST("/v1/models/*action", geminiHandler.NativeAPI)
	router.POST("/v1beta/models/*action", geminiHandler.NativeAPI)

	api := router.Group("/api")
	{
		api.POST("/auth/login", authHandler.Login)
		api.GET("/auth/me", requireLogin(authService), authHandler.Me)

		meGroup := api.Group("/me", requireLogin(authService))
		{
			meGroup.GET("/models", meHandler.ListModels)
			meGroup.GET("/api-keys", meHandler.ListAPIKeys)
			meGroup.POST("/api-keys", meHandler.CreateAPIKey)
			meGroup.GET("/api-keys/:id", meHandler.GetAPIKey)
			meGroup.POST("/api-keys/:id/reveal", meHandler.RevealAPIKey)
			meGroup.PUT("/api-keys/:id", meHandler.UpdateAPIKey)
			meGroup.DELETE("/api-keys/:id", meHandler.DeleteAPIKey)
			meGroup.GET("/api-keys/:id/model-stats", meHandler.ListAPIKeyModelStats)
			meGroup.GET("/api-keys/:id/logs", meHandler.ListAPIKeyLogs)
			meGroup.POST("/api-keys/:id/debug/messages", meHandler.DebugMessages)
			meGroup.POST("/api-keys/:id/debug/chat/completions", meHandler.DebugChatCompletions)
			meGroup.POST("/api-keys/:id/debug/gemini/generate-content", meHandler.DebugGeminiGenerateContent)
			meGroup.POST("/api-keys/:id/debug/gemini/stream-generate-content", meHandler.DebugGeminiStreamGenerateContent)
			meGroup.GET("/logs/:id", meHandler.GetLog)
		}

		adminGroup := api.Group("/admin", requireRole(authService, "admin"))
		{
			adminGroup.GET("/stats", adminHandler.GetStats)
			adminGroup.GET("/claude-proxies", adminHandler.ListClaudeProxyStatus)
			adminGroup.POST("/claude-proxies/:id/probe", adminHandler.ProbeClaudeProxy)
			adminGroup.POST("/claude-proxies/:id/refresh", adminHandler.RefreshClaudeProxy)

			adminGroup.GET("/providers", adminHandler.ListProviders)
			adminGroup.POST("/providers", adminHandler.CreateProvider)
			adminGroup.GET("/providers/:id", adminHandler.GetProvider)
			adminGroup.GET("/providers/:id/usage-stats", adminHandler.GetProviderUsageStats)
			adminGroup.GET("/providers/:id/provider-models", adminHandler.ListProviderModelsByProvider)
			adminGroup.POST("/providers/:id/provider-models", adminHandler.CreateProviderModelForProvider)
			adminGroup.PUT("/providers/:id/provider-models/:providerModelID", adminHandler.UpdateProviderModelForProvider)
			adminGroup.DELETE("/providers/:id/provider-models/:providerModelID", adminHandler.DeleteProviderModelForProvider)
			adminGroup.PUT("/providers/:id", adminHandler.UpdateProvider)
			adminGroup.DELETE("/providers/:id", adminHandler.DeleteProvider)

			adminGroup.GET("/models", adminHandler.ListModels)
			adminGroup.POST("/models", adminHandler.CreateModel)
			adminGroup.GET("/models/:id", adminHandler.GetModel)
			adminGroup.PUT("/models/:id", adminHandler.UpdateModel)
			adminGroup.DELETE("/models/:id", adminHandler.DeleteModel)
			adminGroup.GET("/model-families", adminHandler.ListModelFamilies)
			adminGroup.GET("/model-families/active", adminHandler.ListActiveModelFamilies)
			adminGroup.POST("/model-families", adminHandler.CreateModelFamily)
			adminGroup.GET("/model-families/:id", adminHandler.GetModelFamily)
			adminGroup.PUT("/model-families/:id", adminHandler.UpdateModelFamily)
			adminGroup.DELETE("/model-families/:id", adminHandler.DeleteModelFamily)
			adminGroup.GET("/models/:id/provider-models", adminHandler.ListProviderModels)
			adminGroup.POST("/models/:id/provider-models", adminHandler.CreateProviderModel)
			adminGroup.PUT("/models/:id/provider-models/:providerModelID", adminHandler.UpdateProviderModel)
			adminGroup.DELETE("/models/:id/provider-models/:providerModelID", adminHandler.DeleteProviderModel)
			adminGroup.POST("/models/:id/provider-models/:providerModelID/activate", adminHandler.SetActiveProviderModel)

			adminGroup.GET("/users", adminHandler.ListUsers)
			adminGroup.POST("/users", adminHandler.CreateUser)
			adminGroup.GET("/users/:id", adminHandler.GetUser)
			adminGroup.PUT("/users/:id", adminHandler.UpdateUser)
			adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)

			adminGroup.GET("/logs", adminHandler.ListLogs)
			adminGroup.GET("/logs/:id", adminHandler.GetLog)
		}
	}

	return router
}

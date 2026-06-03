package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

const logBodyCleanupBatchSize = 200

type RequestLogBodyStore struct {
	rootDir   string
	retention time.Duration
}

func NewRequestLogBodyStore(rootDir string, retentionDays int) *RequestLogBodyStore {
	if retentionDays <= 0 {
		retentionDays = 7
	}
	return &RequestLogBodyStore{
		rootDir:   filepath.Clean(rootDir),
		retention: time.Duration(retentionDays) * 24 * time.Hour,
	}
}

func (s *RequestLogBodyStore) Save(ctx context.Context, logs *repository.RequestLogRepository, item *entity.RequestLog, requestBody []byte, responseBody []byte) {
	if s == nil || item == nil || item.ID == 0 {
		return
	}
	if len(requestBody) == 0 && len(responseBody) == 0 {
		return
	}
	createdAt := time.Now().UTC()
	if !item.CreatedAt.IsZero() {
		createdAt = item.CreatedAt.UTC()
	}
	dir := filepath.Join(s.rootDir, createdAt.Format("2006-01-02"))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		log.Printf("create request log body dir: %v", err)
		return
	}

	requestPath := ""
	responsePath := ""
	var err error
	if len(requestBody) > 0 {
		requestPath, err = s.writeBodyFile(dir, item.RequestID, "request", requestBody)
		if err != nil {
			log.Printf("write request log request body: %v", err)
		}
	}
	if len(responseBody) > 0 {
		responsePath, err = s.writeBodyFile(dir, item.RequestID, "response", responseBody)
		if err != nil {
			log.Printf("write request log response body: %v", err)
		}
	}
	if requestPath == "" && responsePath == "" {
		return
	}

	bodyItem := &entity.RequestLogBody{
		RequestLogID:     item.ID,
		RequestBodyPath:  requestPath,
		ResponseBodyPath: responsePath,
		RequestBodySize:  int64(len(requestBody)),
		ResponseBodySize: int64(len(responseBody)),
		ExpiresAt:        createdAt.Add(s.retention),
	}
	if err := logs.CreateBody(ctx, bodyItem); err != nil {
		log.Printf("create request log body index: %v", err)
	}
}

func (s *RequestLogBodyStore) Attach(ctx context.Context, logs *repository.RequestLogRepository, item *entity.RequestLog) {
	if s == nil || item == nil || item.ID == 0 {
		return
	}
	body, err := logs.GetBodyByRequestLogID(ctx, item.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if !item.CreatedAt.IsZero() && time.Now().UTC().After(item.CreatedAt.UTC().Add(s.retention)) {
				s.attachExpiredMessage(item, "原始请求体已超过 7 天并清理", "原始响应体已超过 7 天并清理")
			}
			return
		}
		s.attachExpiredMessage(item, "原始请求体读取失败", "原始响应体读取失败")
		return
	}
	if time.Now().UTC().After(body.ExpiresAt.UTC()) {
		s.attachExpiredMessage(item, "原始请求体已超过 7 天并清理", "原始响应体已超过 7 天并清理")
		return
	}

	requestBody := s.readBody(body.RequestBodyPath, "原始请求体文件不存在")
	responseBody := s.readBody(body.ResponseBodyPath, "原始响应体文件不存在")
	item.RequestPreview = mustJSON(map[string]any{"body": requestBody})
	item.ResponsePreview = mustJSON(map[string]any{"body": responseBody})
}

func (s *RequestLogBodyStore) CleanupExpired(ctx context.Context, logs *repository.RequestLogRepository) {
	if s == nil {
		return
	}
	for {
		items, err := logs.ListExpiredBodies(ctx, time.Now().UTC(), logBodyCleanupBatchSize)
		if err != nil {
			log.Printf("list expired request log bodies: %v", err)
			return
		}
		if len(items) == 0 {
			return
		}
		for _, item := range items {
			removeLogBodyFile(item.RequestBodyPath)
			removeLogBodyFile(item.ResponseBodyPath)
			if err := logs.DeleteBody(ctx, item.ID); err != nil {
				log.Printf("delete request log body index: %v", err)
			}
		}
		if len(items) < logBodyCleanupBatchSize {
			return
		}
	}
}

func (s *RequestLogBodyStore) StartCleanupLoop(ctx context.Context, logs *repository.RequestLogRepository) {
	go func() {
		s.CleanupExpired(ctx, logs)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.CleanupExpired(ctx, logs)
			}
		}
	}()
}

func (s *RequestLogBodyStore) writeBodyFile(dir string, requestID string, kind string, body []byte) (string, error) {
	filename := fmt.Sprintf("%s-%s.log", requestID, kind)
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, body, 0o640); err != nil {
		return "", err
	}
	return path, nil
}

func (s *RequestLogBodyStore) readBody(path string, fallback string) string {
	if path == "" {
		return fallback
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return string(body)
}

func (s *RequestLogBodyStore) attachExpiredMessage(item *entity.RequestLog, requestMessage string, responseMessage string) {
	item.RequestPreview = mustJSON(map[string]any{"expired": true, "message": requestMessage})
	item.ResponsePreview = mustJSON(map[string]any{"expired": true, "message": responseMessage})
}

func removeLogBodyFile(path string) {
	if path == "" {
		return
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("delete request log body file %s: %v", path, err)
	}
}

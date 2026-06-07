package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hrbackend/internal/domain"

	"github.com/gin-gonic/gin"
)

type fakeScheduleService struct {
	domain.ScheduleService
	stats *domain.ShiftSwapStats
	err   error
}

func (f *fakeScheduleService) GetShiftSwapStats(ctx context.Context) (*domain.ShiftSwapStats, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.stats, nil
}

func TestShiftSwapHandlerGetStatsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedStats := &domain.ShiftSwapStats{
		WaitingResponseCount: 3,
		WaitingApprovalCount: 5,
		HandledCount:         12,
	}

	router := gin.New()
	handler := NewShiftSwapHandler(&fakeScheduleService{stats: expectedStats})
	router.GET("/shift-swaps/stats", handler.GetShiftSwapStats)

	req := httptest.NewRequest(http.MethodGet, "/shift-swaps/stats", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    shiftSwapStatsResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !response.Success {
		t.Fatalf("expected success response")
	}
	if response.Message != "Shift swap stats retrieved successfully" {
		t.Fatalf("unexpected message: %s", response.Message)
	}
	if response.Data.WaitingResponseCount != 3 {
		t.Fatalf("expected waiting response count 3, got %d", response.Data.WaitingResponseCount)
	}
	if response.Data.WaitingApprovalCount != 5 {
		t.Fatalf("expected waiting approval count 5, got %d", response.Data.WaitingApprovalCount)
	}
	if response.Data.HandledCount != 12 {
		t.Fatalf("expected handled count 12, got %d", response.Data.HandledCount)
	}
}

func TestShiftSwapHandlerGetStatsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewShiftSwapHandler(&fakeScheduleService{err: errors.New("database connection down")})
	router.GET("/shift-swaps/stats", handler.GetShiftSwapStats)

	req := httptest.NewRequest(http.MethodGet, "/shift-swaps/stats", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.Success {
		t.Fatalf("expected failure response")
	}
	if response.Message != "database connection down" {
		t.Fatalf("unexpected error message: %s", response.Message)
	}
}

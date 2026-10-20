package repositories_test

import (
	"testing"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/repositories"
)

func TestAIModelUsageRepositoryUpsertAndFindAll(t *testing.T) {
	repo := repositories.NewAIModelUsageRepository(newTestDB(t))

	loc, _ := time.LoadLocation("America/Los_Angeles")
	resetDay := time.Date(2026, 10, 20, 0, 0, 0, 0, loc)
	resetMin := time.Date(2026, 10, 19, 12, 31, 0, 0, time.UTC)

	usage := &models.AIModelUsage{Name: "model-a", UsedRequestsDay: 3, UsedRequestsMin: 1, NextResetDay: resetDay, NextResetMin: resetMin}
	if err := repo.Upsert(usage); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	usage.UsedRequestsDay = 4
	usage.UsedRequestsMin = 2
	if err := repo.Upsert(usage); err != nil {
		t.Fatalf("Upsert() second call error = %v", err)
	}

	got, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 usage row, got %d: %+v", len(got), got)
	}
	if got[0].UsedRequestsDay != 4 || got[0].UsedRequestsMin != 2 {
		t.Errorf("expected updated counters 4/2, got %d/%d", got[0].UsedRequestsDay, got[0].UsedRequestsMin)
	}
	if !got[0].NextResetDay.Equal(resetDay) {
		t.Errorf("NextResetDay = %v, want %v", got[0].NextResetDay, resetDay)
	}
	if !got[0].NextResetMin.Equal(resetMin) {
		t.Errorf("NextResetMin = %v, want %v", got[0].NextResetMin, resetMin)
	}
}

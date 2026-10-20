package chatter

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/services"
)

type fakeUsageRepo struct {
	mu    sync.Mutex
	saved int
}

func (r *fakeUsageRepo) FindAll() ([]*models.AIModelUsage, error) { return nil, nil }
func (r *fakeUsageRepo) Close() error                              { return nil }
func (r *fakeUsageRepo) Upsert(*models.AIModelUsage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saved++
	return nil
}

func TestAcquireModelDoesNotExceedLimitConcurrently(t *testing.T) {
	original := Models
	t.Cleanup(func() { Models = original })
	Models = []*Model{NewModel("model-a", 100, 3, 1000)}

	repo := &fakeUsageRepo{}
	service := services.NewAIModelUsageService(repo)

	var acquired atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if _, err := AcquireModel(service); err == nil {
				acquired.Add(1)
			}
		})
	}
	wg.Wait()

	if got := acquired.Load(); got != 3 {
		t.Errorf("expected exactly 3 acquisitions with a per-minute limit of 3, got %d", got)
	}
	if repo.saved != 3 {
		t.Errorf("expected usage saved 3 times, got %d", repo.saved)
	}
}

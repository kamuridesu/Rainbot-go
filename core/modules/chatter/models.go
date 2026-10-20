package chatter

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/services"
)

type Model struct {
	Name              string
	RequestsPerDay    int
	RequestsPerMinute int
	TokensPerMinute   int

	NextResetDay time.Time
	NextResetMin time.Time

	UsedRequestsDay int
	UsedRequestsMin int
}

func NewModel(name string, rpd, rpm, tpm int) *Model {
	m := &Model{
		Name:              name,
		RequestsPerDay:    rpd,
		RequestsPerMinute: rpm,
		TokensPerMinute:   tpm,
	}
	m.resetDay()
	m.resetMinute()
	return m
}

func (m *Model) resetDay() {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	now := time.Now().In(loc)
	m.NextResetDay = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	m.UsedRequestsDay = 0
}

func (m *Model) resetMinute() {
	now := time.Now()
	m.NextResetMin = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	m.UsedRequestsMin = 0
}

func (m *Model) CheckAndResetIfNeeded() {
	now := time.Now()

	if now.After(m.NextResetDay) {
		m.resetDay()
	}

	if now.After(m.NextResetMin) {
		m.resetMinute()
	}
}

func (m *Model) RegisterRequest() error {
	m.CheckAndResetIfNeeded()

	if m.UsedRequestsDay >= m.RequestsPerDay {
		return fmt.Errorf("limite diário excedido no modelo %s", m.Name)
	}
	if m.UsedRequestsMin >= m.RequestsPerMinute {
		return fmt.Errorf("limite por minuto excedido no modelo %s. Tente em %v", m.Name, time.Until(m.NextResetMin))
	}

	m.UsedRequestsDay++
	m.UsedRequestsMin++
	return nil
}

func (m *Model) HasRequests() bool {
	m.CheckAndResetIfNeeded()
	return m.UsedRequestsDay < m.RequestsPerDay && m.UsedRequestsMin < m.RequestsPerMinute
}

var (
	Models = []*Model{
		NewModel("gemini-3.5-flash-lite", 500, 15, 250000),
		NewModel("gemini-3.1-flash-lite", 500, 15, 250000),
		NewModel("gemini-3.8-flash", 20, 5, 250000),
		NewModel("gemini-3.6-flash", 20, 5, 250000),
		NewModel("gemini-3-flash", 20, 5, 250000),
		NewModel("gemma-4-31b-it", 14400, 30, 16000),
		NewModel("gemma-4-26b-it", 14400, 30, 16000),
	}
)

var mu sync.Mutex

func (m *Model) toUsage() *models.AIModelUsage {
	return &models.AIModelUsage{
		Name:            m.Name,
		UsedRequestsDay: m.UsedRequestsDay,
		UsedRequestsMin: m.UsedRequestsMin,
		NextResetDay:    m.NextResetDay,
		NextResetMin:    m.NextResetMin,
	}
}

func LoadUsage(service *services.AIModelUsageService) error {
	usages, err := service.GetAll()
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()
	for _, usage := range usages {
		for _, model := range Models {
			if model.Name != usage.Name {
				continue
			}
			model.UsedRequestsDay = usage.UsedRequestsDay
			model.UsedRequestsMin = usage.UsedRequestsMin
			model.NextResetDay = usage.NextResetDay
			model.NextResetMin = usage.NextResetMin
			model.CheckAndResetIfNeeded()
		}
	}
	return nil
}

func AcquireModel(service *services.AIModelUsageService) (*Model, error) {
	mu.Lock()
	var acquired *Model
	var usage *models.AIModelUsage
	for _, model := range Models {
		if model.HasRequests() && model.RegisterRequest() == nil {
			acquired = model
			usage = model.toUsage()
			break
		}
	}
	mu.Unlock()

	if acquired == nil {
		return nil, errors.New("no models available")
	}

	slog.Info(fmt.Sprintf("Model %s acquired, daily requests left: %d", acquired.Name, acquired.RequestsPerDay-usage.UsedRequestsDay))
	if err := service.Save(usage); err != nil {
		slog.Error(fmt.Sprintf("failed to save usage for model %s: %v", acquired.Name, err))
	}
	return acquired, nil
}

package models

import "time"

type AIModelUsage struct {
	Name            string    `json:"name"`
	UsedRequestsDay int       `json:"usedRequestsDay"`
	UsedRequestsMin int       `json:"usedRequestsMin"`
	NextResetDay    time.Time `json:"nextResetDay"`
	NextResetMin    time.Time `json:"nextResetMin"`
}

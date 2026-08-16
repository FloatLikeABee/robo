package models

import (
	"time"
)

// EventInfo is stored in Badger (former Mongo workspace_events collection).
type EventInfo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	Reporter  string    `json:"reporter"`
	EventTime time.Time `json:"time"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateEventInfoRequest struct {
	Title    string `json:"title" binding:"required"`
	Detail   string `json:"detail"`
	Reporter string `json:"reporter"`
	// Time is RFC3339 / ISO8601 (e.g. 2006-01-02T15:04:05Z07:00).
	Time string `json:"time" binding:"required"`
}

type EventInfoAIContextResponse struct {
	Event *EventInfoResponse `json:"event"`
}

type EventInfoResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	Reporter  string    `json:"reporter"`
	Time      time.Time `json:"time"`
	CreatedAt time.Time `json:"created_at"`
}

func EventInfoToResponse(e *EventInfo) *EventInfoResponse {
	if e == nil {
		return nil
	}
	return &EventInfoResponse{
		ID:        e.ID,
		Title:     e.Title,
		Detail:    e.Detail,
		Reporter:  e.Reporter,
		Time:      e.EventTime,
		CreatedAt: e.CreatedAt,
	}
}

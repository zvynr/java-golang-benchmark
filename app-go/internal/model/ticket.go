package model

import "time"

type Ticket struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Assignee    string    `json:"assignee"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateTicketRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Priority    string `json:"priority" binding:"required"`
	Assignee    string `json:"assignee"`
}

type TicketStatistics struct {
	Total      int64 `json:"total"`
	Open       int64 `json:"open"`
	InProgress int64 `json:"inProgress"`
	Resolved   int64 `json:"resolved"`
	Closed     int64 `json:"closed"`
	Critical   int64 `json:"critical"`
	High       int64 `json:"high"`
	Medium     int64 `json:"medium"`
	Low        int64 `json:"low"`
}

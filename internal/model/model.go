package model

import "time"

type StandardResponse[T any] struct {
	Success   bool      `json:"success"`
	Data      T         `json:"data"`
	Message   string    `json:"message"`
	Meta      *Meta     `json:"meta,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type Meta struct {
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func NewSuccessResponse[T any](data T, message string) *StandardResponse[T] {
	return &StandardResponse[T]{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func NewSuccessResponseWithMeta[T any](data T, message string, meta *Meta) *StandardResponse[T] {
	return &StandardResponse[T]{
		Success:   true,
		Data:      data,
		Message:   message,
		Meta:      meta,
		Timestamp: time.Now(),
	}
}

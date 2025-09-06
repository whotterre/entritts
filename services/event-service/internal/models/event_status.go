package models

import (
	"database/sql/driver"
	"errors"
)

type EventStatus string

const (
	EventStatusDraft     EventStatus = "DRAFT"
	EventStatusPublished EventStatus = "PUBLISHED"
	EventStatusCancelled EventStatus = "CANCELLED"
	EventStatusCompleted EventStatus = "COMPLETED"
)

// Implement SQL Scanner interface
func (es *EventStatus) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*es = EventStatus(v)
	case string:
		*es = EventStatus(v)
	default:
		return errors.New("invalid event status type")
	}
	return nil
}

// Implement SQL Valuer interface
func (es EventStatus) Value() (driver.Value, error) {
	return string(es), nil
}

package utils

import "github.com/google/uuid"

// StringToUUIDFormat converts a string to UUID format.
// Returns empty UUID if the string is invalid.
func StringToUUIDFormat(text string) uuid.UUID {
	id, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}
	}
	return id
}
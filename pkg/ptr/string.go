package ptr

import (
	"strings"

	"github.com/google/uuid"
)

func String(value string) *string {
	return &value
}

func OrEmpty(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func TrimString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func UUIDOrNil(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

package chat

import (
	"time"

	"github.com/google/uuid"
)

func NewID() string {
	return uuid.New().String()
}

func Now() time.Time {
	return time.Now()
}

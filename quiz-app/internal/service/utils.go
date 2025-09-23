package service

import (
	"fmt"
	"time"
)

// generateSessionID generates a simple session ID without external dependencies
func generateSessionID() string {
	return fmt.Sprintf("session_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

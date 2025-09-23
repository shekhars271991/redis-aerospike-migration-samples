package service

import (
	"testing"
)

func TestGenerateSessionID(t *testing.T) {
	sessionID := generateSessionID()
	if sessionID == "" {
		t.Error("generateSessionID should not return empty string")
	}
	
	// Check that two consecutive calls return different IDs
	sessionID2 := generateSessionID()
	if sessionID == sessionID2 {
		t.Error("generateSessionID should return unique IDs")
	}
}

// Note: Integration tests would require Redis/Aerospike instances
// For production, you would add comprehensive integration tests here

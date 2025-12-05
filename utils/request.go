package utils

import (
	"net/http"
	"strconv"
)

// GetUserIDFromRequest extracts user ID from request headers (set by middleware)
func GetUserIDFromRequest(r *http.Request) int {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0
	}
	userID, _ := strconv.Atoi(userIDStr)
	return userID
}


package handlers

import (
	"codesprint/database"
	"codesprint/models"
	"codesprint/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// CreateContest handles contest creation (admin only)
func CreateContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := utils.GetUserIDFromRequest(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateContestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if req.EndTime.Before(req.StartTime) {
		http.Error(w, "End time must be after start time", http.StatusBadRequest)
		return
	}

	// Create contest
	var contestID int
	err := database.DB.QueryRow(
		"INSERT INTO contests (title, start_time, end_time, created_by) VALUES ($1, $2, $3, $4) RETURNING id",
		req.Title, req.StartTime, req.EndTime, userID,
	).Scan(&contestID)
	if err != nil {
		http.Error(w, "Failed to create contest", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         contestID,
		"title":      req.Title,
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
	})
}

// GetContests returns all contests
func GetContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, title, start_time, end_time, created_by, created_at 
		FROM contests 
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, "Failed to fetch contests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var contests []models.Contest
	for rows.Next() {
		var contest models.Contest
		err := rows.Scan(&contest.ID, &contest.Title, &contest.StartTime, &contest.EndTime, &contest.CreatedBy, &contest.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan contest", http.StatusInternalServerError)
			return
		}
		contests = append(contests, contest)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contests)
}

// GetContest returns a specific contest
func GetContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get contest ID from URL path (mux variable)
	vars := mux.Vars(r)
	contestIDStr := vars["id"]
	if contestIDStr == "" {
		contestIDStr = r.URL.Query().Get("id")
	}
	contestID, err := strconv.Atoi(contestIDStr)
	if err != nil {
		http.Error(w, "Invalid contest ID", http.StatusBadRequest)
		return
	}

	var contest models.Contest
	err = database.DB.QueryRow(
		"SELECT id, title, start_time, end_time, created_by, created_at FROM contests WHERE id = $1",
		contestID,
	).Scan(&contest.ID, &contest.Title, &contest.StartTime, &contest.EndTime, &contest.CreatedBy, &contest.CreatedAt)
	if err != nil {
		http.Error(w, "Contest not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contest)
}


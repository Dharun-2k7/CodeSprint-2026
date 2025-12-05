package handlers

import (
	"codesprint/database"
	"codesprint/models"
	"codesprint/utils"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// CreateProblem handles problem creation (admin only)
func CreateProblem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := utils.GetUserIDFromRequest(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Title == "" || req.Statement == "" {
		http.Error(w, "Title and statement are required", http.StatusBadRequest)
		return
	}

	if req.TimeLimit <= 0 {
		req.TimeLimit = 1000 // default 1 second
	}
	if req.MemoryLimit <= 0 {
		req.MemoryLimit = 256 // default 256 MB
	}

	// Create problem
	var problemID int
	err := database.DB.QueryRow(
		"INSERT INTO problems (contest_id, title, statement, time_limit, memory_limit) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		req.ContestID, req.Title, req.Statement, req.TimeLimit, req.MemoryLimit,
	).Scan(&problemID)
	if err != nil {
		http.Error(w, "Failed to create problem", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           problemID,
		"contest_id":   req.ContestID,
		"title":        req.Title,
		"statement":    req.Statement,
		"time_limit":   req.TimeLimit,
		"memory_limit": req.MemoryLimit,
	})
}

// GetContestProblems returns all problems for a contest
func GetContestProblems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contestIDStr := r.URL.Query().Get("contest_id")
	contestID, err := strconv.Atoi(contestIDStr)
	if err != nil {
		http.Error(w, "Invalid contest ID", http.StatusBadRequest)
		return
	}

	rows, err := database.DB.Query(
		"SELECT id, contest_id, title, statement, time_limit, memory_limit, created_at FROM problems WHERE contest_id = $1 ORDER BY id",
		contestID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch problems", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var problems []models.Problem
	for rows.Next() {
		var problem models.Problem
		err := rows.Scan(&problem.ID, &problem.ContestID, &problem.Title, &problem.Statement, &problem.TimeLimit, &problem.MemoryLimit, &problem.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan problem", http.StatusInternalServerError)
			return
		}
		problems = append(problems, problem)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problems)
}

// GetProblem returns a specific problem
func GetProblem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get problem ID from URL path (mux variable)
	vars := mux.Vars(r)
	problemIDStr := vars["id"]
	if problemIDStr == "" {
		problemIDStr = r.URL.Query().Get("id")
	}
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		http.Error(w, "Invalid problem ID", http.StatusBadRequest)
		return
	}

	var problem models.Problem
	err = database.DB.QueryRow(
		"SELECT id, contest_id, title, statement, time_limit, memory_limit, created_at FROM problems WHERE id = $1",
		problemID,
	).Scan(&problem.ID, &problem.ContestID, &problem.Title, &problem.Statement, &problem.TimeLimit, &problem.MemoryLimit, &problem.CreatedAt)
	if err != nil {
		http.Error(w, "Problem not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problem)
}


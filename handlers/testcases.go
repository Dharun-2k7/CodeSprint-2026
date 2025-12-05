package handlers

import (
	"codesprint/database"
	"codesprint/models"
	"codesprint/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

// CreateTestcase handles testcase creation (admin only)
func CreateTestcase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := utils.GetUserIDFromRequest(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ProblemID      int    `json:"problem_id"`
		Input          string `json:"input"`
		ExpectedOutput string `json:"expected_output"`
		IsSample       bool   `json:"is_sample"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Input == "" || req.ExpectedOutput == "" {
		http.Error(w, "Input and expected_output are required", http.StatusBadRequest)
		return
	}

	// Create testcase
	var testcaseID int
	err := database.DB.QueryRow(
		"INSERT INTO testcases (problem_id, input, expected_output, is_sample) VALUES ($1, $2, $3, $4) RETURNING id",
		req.ProblemID, req.Input, req.ExpectedOutput, req.IsSample,
	).Scan(&testcaseID)
	if err != nil {
		http.Error(w, "Failed to create testcase", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             testcaseID,
		"problem_id":     req.ProblemID,
		"is_sample":      req.IsSample,
	})
}

// GetTestcases returns testcases for a problem (only sample ones for users)
func GetTestcases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	problemIDStr := r.URL.Query().Get("problem_id")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		http.Error(w, "Invalid problem ID", http.StatusBadRequest)
		return
	}

	// For MVP, only return sample testcases
	rows, err := database.DB.Query(
		"SELECT id, problem_id, input, expected_output, is_sample FROM testcases WHERE problem_id = $1 AND is_sample = true ORDER BY id",
		problemID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch testcases", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var testcases []models.Testcase
	for rows.Next() {
		var tc models.Testcase
		err := rows.Scan(&tc.ID, &tc.ProblemID, &tc.Input, &tc.ExpectedOutput, &tc.IsSample)
		if err != nil {
			continue
		}
		testcases = append(testcases, tc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testcases)
}


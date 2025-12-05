package handlers

import (
	"codesprint/database"
	"codesprint/judge"
	"codesprint/models"
	"codesprint/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// SubmitCode handles code submission
func SubmitCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := utils.GetUserIDFromRequest(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.SubmitCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Code == "" || req.Language == "" {
		http.Error(w, "Code and language are required", http.StatusBadRequest)
		return
	}

	// Get problem to check time limit
	var problem models.Problem
	err := database.DB.QueryRow(
		"SELECT id, contest_id, title, time_limit, memory_limit FROM problems WHERE id = $1",
		req.ProblemID,
	).Scan(&problem.ID, &problem.ContestID, &problem.Title, &problem.TimeLimit, &problem.MemoryLimit)
	if err != nil {
		http.Error(w, "Problem not found", http.StatusNotFound)
		return
	}

	// Get all testcases for the problem
	rows, err := database.DB.Query(
		"SELECT id, input, expected_output FROM testcases WHERE problem_id = $1",
		req.ProblemID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch testcases", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var testcases []models.Testcase
	for rows.Next() {
		var tc models.Testcase
		err := rows.Scan(&tc.ID, &tc.Input, &tc.ExpectedOutput)
		if err != nil {
			continue
		}
		testcases = append(testcases, tc)
	}

	if len(testcases) == 0 {
		http.Error(w, "No testcases found for this problem", http.StatusBadRequest)
		return
	}

	// Create submission record
	var submissionID int
	err = database.DB.QueryRow(
		"INSERT INTO submissions (user_id, problem_id, contest_id, language, code, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		userID, req.ProblemID, req.ContestID, req.Language, req.Code, "pending",
	).Scan(&submissionID)
	if err != nil {
		http.Error(w, "Failed to create submission", http.StatusInternalServerError)
		return
	}

	// Process submission asynchronously
	go processSubmission(submissionID, req.Code, req.Language, testcases, problem.TimeLimit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"submission_id": submissionID,
		"status":        "pending",
	})
}

// processSubmission processes a submission against all testcases
func processSubmission(submissionID int, code, language string, testcases []models.Testcase, timeLimit int) {
	languageID := judge.GetLanguageID(language)
	allPassed := true
	totalRuntime := 0
	finalStatus := "accepted"

	// Process each testcase
	for _, tc := range testcases {
		// Submit to Judge0
		result, err := judge.SubmitCode(code, languageID, tc.Input)
		if err != nil {
			finalStatus = "runtime_error"
			allPassed = false
			break
		}

		// Poll for result
		pollResult, err := judge.PollSubmissionResult(result.Token, 30, time.Second*2)
		if err != nil {
			finalStatus = "runtime_error"
			allPassed = false
			break
		}

		// Parse runtime
		if pollResult.Time != "" {
			// Judge0 returns time as "0.001" (seconds), convert to milliseconds
			var runtimeSeconds float64
			fmt.Sscanf(pollResult.Time, "%f", &runtimeSeconds)
			runtimeMs := int(runtimeSeconds * 1000)
			if runtimeMs > totalRuntime {
				totalRuntime = runtimeMs
			}
		}

		// Check result status
		status := judge.MapJudge0StatusToInternal(pollResult.Status.ID)
		
		// Check if output matches (only if Judge0 says accepted)
		if pollResult.Status.ID == 3 { // Judge0 accepted status
			// Trim whitespace for comparison
			output := trimWhitespace(pollResult.Stdout)
			expected := trimWhitespace(tc.ExpectedOutput)
			if output != expected {
				status = "wrong_answer"
			}
		}

		if status != "accepted" {
			finalStatus = status
			allPassed = false
			break
		}
	}

	// Calculate score
	score := 0
	if allPassed {
		score = 100
	}

	// Update submission
	_, err := database.DB.Exec(
		"UPDATE submissions SET status = $1, score = $2, runtime = $3 WHERE id = $4",
		finalStatus, score, totalRuntime, submissionID,
	)
	if err != nil {
		fmt.Printf("Failed to update submission %d: %v\n", submissionID, err)
	}

	// Update leaderboard cache
	updateLeaderboardCache(submissionID)
}

func trimWhitespace(s string) string {
	// Trim leading and trailing whitespace, normalize line endings
	lines := []string{}
	currentLine := ""
	for _, char := range s {
		if char == '\n' || char == '\r' {
			if len(currentLine) > 0 {
				lines = append(lines, currentLine)
				currentLine = ""
			}
		} else {
			currentLine += string(char)
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}
	// Join lines with newline
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

// GetSubmission returns a submission by ID
func GetSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get submission ID from URL path (mux variable)
	vars := mux.Vars(r)
	submissionIDStr := vars["id"]
	if submissionIDStr == "" {
		submissionIDStr = r.URL.Query().Get("id")
	}
	submissionID, err := strconv.Atoi(submissionIDStr)
	if err != nil {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	var submission models.Submission
	err = database.DB.QueryRow(
		"SELECT id, user_id, problem_id, contest_id, language, code, status, score, runtime, created_at FROM submissions WHERE id = $1",
		submissionID,
	).Scan(&submission.ID, &submission.UserID, &submission.ProblemID, &submission.ContestID, &submission.Language, &submission.Code, &submission.Status, &submission.Score, &submission.Runtime, &submission.CreatedAt)
	if err != nil {
		http.Error(w, "Submission not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submission)
}

// GetUserSubmissions returns all submissions for a user in a contest
func GetUserSubmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := utils.GetUserIDFromRequest(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	contestIDStr := r.URL.Query().Get("contest_id")
	contestID, err := strconv.Atoi(contestIDStr)
	if err != nil {
		http.Error(w, "Invalid contest ID", http.StatusBadRequest)
		return
	}

	rows, err := database.DB.Query(
		"SELECT id, user_id, problem_id, contest_id, language, status, score, runtime, created_at FROM submissions WHERE user_id = $1 AND contest_id = $2 ORDER BY created_at DESC",
		userID, contestID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch submissions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var submissions []models.Submission
	for rows.Next() {
		var sub models.Submission
		err := rows.Scan(&sub.ID, &sub.UserID, &sub.ProblemID, &sub.ContestID, &sub.Language, &sub.Status, &sub.Score, &sub.Runtime, &sub.CreatedAt)
		if err != nil {
			continue
		}
		submissions = append(submissions, sub)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}

// updateLeaderboardCache updates the leaderboard cache for a submission
func updateLeaderboardCache(submissionID int) {
	var submission models.Submission
	err := database.DB.QueryRow(
		"SELECT user_id, contest_id, problem_id, status, created_at FROM submissions WHERE id = $1",
		submissionID,
	).Scan(&submission.UserID, &submission.ContestID, &submission.ProblemID, &submission.Status, &submission.CreatedAt)
	if err != nil {
		return
	}

	// Only count accepted submissions
	if submission.Status != "accepted" {
		return
	}

	// Check if this is the first accepted submission for this problem
	var existingCount int
	err = database.DB.QueryRow(
		"SELECT COUNT(*) FROM submissions WHERE user_id = $1 AND contest_id = $2 AND problem_id = $3 AND status = 'accepted'",
		submission.UserID, submission.ContestID, submission.ProblemID,
	).Scan(&existingCount)
	if err != nil || existingCount > 1 {
		// Not the first accepted, don't update
		return
	}

	// Get current leaderboard entry
	var solvedCount, penalty int
	var lastSubmissionTime time.Time
	err = database.DB.QueryRow(
		"SELECT solved_count, penalty, last_submission_time FROM leaderboard_cache WHERE contest_id = $1 AND user_id = $2",
		submission.ContestID, submission.UserID,
	).Scan(&solvedCount, &penalty, &lastSubmissionTime)

	if err != nil {
		// Create new entry
		contestStart, _ := getContestStartTime(submission.ContestID)
		penaltyMinutes := int(submission.CreatedAt.Sub(contestStart).Minutes())
		database.DB.Exec(
			"INSERT INTO leaderboard_cache (contest_id, user_id, solved_count, penalty, last_submission_time) VALUES ($1, $2, $3, $4, $5)",
			submission.ContestID, submission.UserID, 1, penaltyMinutes, submission.CreatedAt,
		)
	} else {
		// Update existing entry
		contestStart, _ := getContestStartTime(submission.ContestID)
		penaltyMinutes := int(submission.CreatedAt.Sub(contestStart).Minutes())
		database.DB.Exec(
			"UPDATE leaderboard_cache SET solved_count = solved_count + 1, penalty = penalty + $1, last_submission_time = $2 WHERE contest_id = $3 AND user_id = $4",
			penaltyMinutes, submission.CreatedAt, submission.ContestID, submission.UserID,
		)
	}
}

func getContestStartTime(contestID int) (time.Time, error) {
	var startTime time.Time
	err := database.DB.QueryRow(
		"SELECT start_time FROM contests WHERE id = $1",
		contestID,
	).Scan(&startTime)
	return startTime, err
}



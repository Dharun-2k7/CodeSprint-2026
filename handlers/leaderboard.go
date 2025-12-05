package handlers

import (
	"codesprint/database"
	"codesprint/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GetLeaderboard returns the leaderboard for a contest
func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get contest ID from URL path (mux variable)
	vars := mux.Vars(r)
	contestIDStr := vars["contest_id"]
	if contestIDStr == "" {
		contestIDStr = r.URL.Query().Get("contest_id")
	}
	contestID, err := strconv.Atoi(contestIDStr)
	if err != nil {
		http.Error(w, "Invalid contest ID", http.StatusBadRequest)
		return
	}

	// Recompute leaderboard from submissions
	rows, err := database.DB.Query(`
		SELECT 
			u.id as user_id,
			u.name as user_name,
			COUNT(DISTINCT CASE WHEN s.status = 'accepted' THEN s.problem_id END) as solved_count,
			COALESCE(SUM(CASE WHEN s.status = 'accepted' THEN 
				EXTRACT(EPOCH FROM (s.created_at - c.start_time)) / 60 
			END), 0)::INTEGER as penalty,
			MAX(CASE WHEN s.status = 'accepted' THEN s.created_at END) as last_submission_time
		FROM users u
		LEFT JOIN submissions s ON u.id = s.user_id AND s.contest_id = $1
		LEFT JOIN contests c ON c.id = $1
		WHERE EXISTS (
			SELECT 1 FROM submissions WHERE user_id = u.id AND contest_id = $1
		)
		GROUP BY u.id, u.name
		ORDER BY solved_count DESC, penalty ASC, last_submission_time ASC
	`, contestID)
	if err != nil {
		http.Error(w, "Failed to fetch leaderboard", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var leaderboard []models.LeaderboardEntry
	for rows.Next() {
		var entry models.LeaderboardEntry
		var lastSubmissionTime *string
		err := rows.Scan(&entry.UserID, &entry.UserName, &entry.SolvedCount, &entry.Penalty, &lastSubmissionTime)
		if err != nil {
			continue
		}
		if lastSubmissionTime != nil {
			// Parse time if needed
		}
		leaderboard = append(leaderboard, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}


package models

import "time"

// User represents a user in the system
type User struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Contest represents a programming contest
type Contest struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`
	CreatedBy int       `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Problem represents a problem in a contest
type Problem struct {
	ID          int       `json:"id" db:"id"`
	ContestID   int       `json:"contest_id" db:"contest_id"`
	Title       string    `json:"title" db:"title"`
	Statement   string    `json:"statement" db:"statement"`
	TimeLimit   int       `json:"time_limit" db:"time_limit"`   // milliseconds
	MemoryLimit int       `json:"memory_limit" db:"memory_limit"` // MB
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Testcase represents a test case for a problem
type Testcase struct {
	ID            int    `json:"id" db:"id"`
	ProblemID     int    `json:"problem_id" db:"problem_id"`
	Input         string `json:"input" db:"input"`
	ExpectedOutput string `json:"expected_output" db:"expected_output"`
	IsSample      bool   `json:"is_sample" db:"is_sample"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Submission represents a code submission
type Submission struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	ProblemID   int       `json:"problem_id" db:"problem_id"`
	ContestID   int       `json:"contest_id" db:"contest_id"`
	Language    string    `json:"language" db:"language"`
	Code        string    `json:"code" db:"code"`
	Status      string    `json:"status" db:"status"`
	Score       int       `json:"score" db:"score"`
	Runtime     int       `json:"runtime" db:"runtime"` // milliseconds
	Judge0Token string    `json:"judge0_token" db:"judge0_token"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	UserID            int       `json:"user_id" db:"user_id"`
	UserName          string    `json:"user_name" db:"user_name"`
	SolvedCount       int       `json:"solved_count" db:"solved_count"`
	Penalty           int       `json:"penalty" db:"penalty"`
	LastSubmissionTime *time.Time `json:"last_submission_time" db:"last_submission_time"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignupRequest represents a signup request
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateContestRequest represents a request to create a contest
type CreateContestRequest struct {
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// CreateProblemRequest represents a request to create a problem
type CreateProblemRequest struct {
	ContestID   int    `json:"contest_id"`
	Title       string `json:"title"`
	Statement   string `json:"statement"`
	TimeLimit   int    `json:"time_limit"`
	MemoryLimit int    `json:"memory_limit"`
}

// SubmitCodeRequest represents a code submission request
type SubmitCodeRequest struct {
	ProblemID int    `json:"problem_id"`
	ContestID int    `json:"contest_id"`
	Language  string `json:"language"`
	Code      string `json:"code"`
}


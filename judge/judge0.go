package judge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var judge0URL = getJudge0URL()

func getJudge0URL() string {
	url := os.Getenv("JUDGE0_URL")
	if url == "" {
		return "http://localhost:2358"
	}
	return url
}

// Language IDs for Judge0
const (
	LanguageC      = 50  // C (GCC 9.2.0)
	LanguageCPP    = 54  // C++ (GCC 9.2.0)
	LanguagePython = 92  // Python (3.8.1)
)

// Judge0Submission represents a submission to Judge0
type Judge0Submission struct {
	SourceCode string `json:"source_code"`
	LanguageID int    `json:"language_id"`
	Stdin      string `json:"stdin,omitempty"`
}

// Judge0Response represents a response from Judge0
type Judge0Response struct {
	Token       string `json:"token"`
	Status      *Judge0Status `json:"status,omitempty"`
	Stdout      string `json:"stdout,omitempty"`
	Stderr      string `json:"stderr,omitempty"`
	Time        string `json:"time,omitempty"`
	Memory      int    `json:"memory,omitempty"`
	CompileOutput string `json:"compile_output,omitempty"`
}

// Judge0Status represents the status of a submission
type Judge0Status struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

// SubmitCode submits code to Judge0
func SubmitCode(code string, languageID int, input string) (*Judge0Response, error) {
	submission := Judge0Submission{
		SourceCode: code,
		LanguageID: languageID,
		Stdin:      input,
	}

	jsonData, err := json.Marshal(submission)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal submission: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/submissions?base64_encoded=false&wait=false", judge0URL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit to Judge0: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Judge0 returned status %d", resp.StatusCode)
	}

	var result Judge0Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetSubmissionResult retrieves the result of a submission from Judge0
func GetSubmissionResult(token string) (*Judge0Response, error) {
	resp, err := http.Get(fmt.Sprintf("%s/submissions/%s?base64_encoded=false", judge0URL, token))
	if err != nil {
		return nil, fmt.Errorf("failed to get result from Judge0: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Judge0 returned status %d", resp.StatusCode)
	}

	var result Judge0Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// PollSubmissionResult polls Judge0 until the submission is complete
func PollSubmissionResult(token string, maxAttempts int, delay time.Duration) (*Judge0Response, error) {
	for i := 0; i < maxAttempts; i++ {
		result, err := GetSubmissionResult(token)
		if err != nil {
			return nil, err
		}

		// Status ID 1-2 means in queue or processing, 3 means completed
		if result.Status != nil && result.Status.ID >= 3 {
			return result, nil
		}

		time.Sleep(delay)
	}

	return nil, fmt.Errorf("submission timed out after %d attempts", maxAttempts)
}

// MapJudge0StatusToInternal maps Judge0 status to internal status
func MapJudge0StatusToInternal(judge0StatusID int) string {
	switch judge0StatusID {
	case 3: // Accepted
		return "accepted"
	case 4: // Wrong Answer
		return "wrong_answer"
	case 5: // Time Limit Exceeded
		return "time_limit_exceeded"
	case 6: // Compilation Error
		return "compilation_error"
	case 7: // Runtime Error
		return "runtime_error"
	case 8: // Memory Limit Exceeded
		return "memory_limit_exceeded"
	default:
		return "pending"
	}
}

// GetLanguageID maps language string to Judge0 language ID
func GetLanguageID(language string) int {
	switch language {
	case "c":
		return LanguageC
	case "cpp", "c++":
		return LanguageCPP
	case "python", "python3":
		return LanguagePython
	default:
		return LanguageC // default to C
	}
}


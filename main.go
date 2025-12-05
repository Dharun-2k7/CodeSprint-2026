package main

import (
	"codesprint/database"
	"codesprint/handlers"
	"codesprint/middleware"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Create router
	r := mux.NewRouter()

	// CORS middleware
	r.Use(corsMiddleware)

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Auth routes (public)
	api.HandleFunc("/signup", handlers.Signup).Methods("POST")
	api.HandleFunc("/login", handlers.Login).Methods("POST")

	// Contest routes
	api.HandleFunc("/contests", handlers.GetContests).Methods("GET")
	api.HandleFunc("/contest/{id:[0-9]+}", handlers.GetContest).Methods("GET")
	api.HandleFunc("/contests", middleware.AuthMiddleware(handlers.CreateContest)).Methods("POST")

	// Problem routes
	api.HandleFunc("/problems", middleware.AuthMiddleware(handlers.GetContestProblems)).Methods("GET")
	api.HandleFunc("/problem/{id:[0-9]+}", handlers.GetProblem).Methods("GET")
	api.HandleFunc("/problems", middleware.AuthMiddleware(handlers.CreateProblem)).Methods("POST")

	// Testcase routes (admin)
	api.HandleFunc("/testcases", middleware.AuthMiddleware(handlers.CreateTestcase)).Methods("POST")
	api.HandleFunc("/testcases", handlers.GetTestcases).Methods("GET")

	// Submission routes
	api.HandleFunc("/submission", middleware.AuthMiddleware(handlers.SubmitCode)).Methods("POST")
	api.HandleFunc("/submission/{id:[0-9]+}", handlers.GetSubmission).Methods("GET")
	api.HandleFunc("/submissions", middleware.AuthMiddleware(handlers.GetUserSubmissions)).Methods("GET")

	// Leaderboard routes
	api.HandleFunc("/leaderboard/{contest_id:[0-9]+}", handlers.GetLeaderboard).Methods("GET")

	// Serve frontend (static files) - must be last to not interfere with API routes
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./frontend/"))))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}


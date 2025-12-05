# Codesprint - Online Judge Platform

An MVP online judge platform for programming contests built with Go, PostgreSQL, and Judge0.

## Features

- **User Authentication**: Sign up and login with JWT-based authentication
- **Contest Management**: Create and manage programming contests
- **Problem Management**: Upload problems with test cases
- **Code Submission**: Submit code in C, C++, and Python3
- **Automatic Judging**: Integration with Judge0 for secure code execution
- **Leaderboard**: Real-time leaderboard with auto-refresh (10-second polling)
- **Responsive UI**: Clean, modern interface built with Bootstrap

## Tech Stack

- **Backend**: Go (net/http, gorilla/mux)
- **Database**: PostgreSQL
- **Judge Engine**: Judge0 (self-hosted via Docker)
- **Frontend**: HTML, CSS, JavaScript (vanilla)
- **Containerization**: Docker & Docker Compose

## Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone <repository-url>
cd CODESPRINT
```

2. Start all services:
```bash
docker-compose up -d
```

This will start:
- PostgreSQL database
- Judge0 (judge engine)
- Backend API server

3. Access the application:
- Frontend: http://localhost:8080
- API: http://localhost:8080/api
- Judge0: http://localhost:2358

4. Wait for services to be ready (especially Judge0, which may take a minute to initialize)

### Local Development

1. Start database and Judge0:
```bash
docker-compose up -d postgres judge0 judge0_db judge0_redis
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Install dependencies:
```bash
go mod download
```

4. Run the application:
```bash
go run main.go
```

## API Endpoints

### Authentication
- `POST /api/signup` - Register a new user
- `POST /api/login` - Login and get JWT token

### Contests
- `GET /api/contests` - List all contests
- `GET /api/contest/{id}` - Get contest details
- `POST /api/contests` - Create a contest (requires auth)

### Problems
- `GET /api/problems?contest_id={id}` - Get problems for a contest
- `GET /api/problem/{id}` - Get problem details
- `POST /api/problems` - Create a problem (requires auth)

### Testcases
- `GET /api/testcases?problem_id={id}` - Get sample testcases
- `POST /api/testcases` - Create a testcase (requires auth)

### Submissions
- `POST /api/submission` - Submit code (requires auth)
- `GET /api/submission/{id}` - Get submission result
- `GET /api/submissions?contest_id={id}` - Get user submissions

### Leaderboard
- `GET /api/leaderboard/{contest_id}` - Get contest leaderboard

## Usage

### Creating a Contest (Admin)

1. Login to the application
2. Use the API to create a contest:
```bash
curl -X POST http://localhost:8080/api/contests \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Summer Contest 2024",
    "start_time": "2024-06-01T10:00:00Z",
    "end_time": "2024-06-01T18:00:00Z"
  }'
```

### Adding Problems

1. Create a problem:
```bash
curl -X POST http://localhost:8080/api/problems \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "contest_id": 1,
    "title": "Hello World",
    "statement": "Print Hello World",
    "time_limit": 1000,
    "memory_limit": 256
  }'
```

2. Add testcases:
```bash
curl -X POST http://localhost:8080/api/testcases \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "problem_id": 1,
    "input": "",
    "expected_output": "Hello World\n",
    "is_sample": true
  }'
```

### Submitting Code

Users can submit code through the web interface or API:
```bash
curl -X POST http://localhost:8080/api/submission \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "problem_id": 1,
    "contest_id": 1,
    "language": "python3",
    "code": "print(\"Hello World\")"
  }'
```

## Supported Languages

- C (GCC 9.2.0)
- C++ (GCC 9.2.0)
- Python 3 (3.8.1)

## Leaderboard Scoring

- **Primary**: Number of problems solved (descending)
- **Secondary**: Total penalty time in minutes (ascending)
- **Tertiary**: Time of last submission (ascending)

Penalty is calculated as the time from contest start to first accepted submission for each problem.

## Security Notes

- Judge0 runs in isolated Docker containers with resource limits
- User code is never executed on the host system
- All user inputs are sanitized
- JWT tokens are used for authentication
- Rate limiting should be added in production

## Project Structure

```
CODESPRINT/
├── database/
│   ├── db.go          # Database connection and initialization
│   └── schema.sql     # Database schema
├── handlers/
│   ├── auth.go        # Authentication handlers
│   ├── contests.go    # Contest management
│   ├── problems.go    # Problem management
│   ├── submissions.go # Submission handling
│   ├── testcases.go   # Testcase management
│   └── leaderboard.go # Leaderboard API
├── judge/
│   └── judge0.go      # Judge0 integration
├── middleware/
│   └── auth.go        # JWT authentication middleware
├── models/
│   └── models.go      # Data models
├── utils/
│   ├── auth.go        # Authentication utilities
│   └── request.go     # Request utilities
├── frontend/
│   ├── index.html     # Main HTML file
│   ├── app.js         # Frontend JavaScript
│   └── styles.css     # Styles
├── main.go            # Application entry point
├── docker-compose.yml # Docker Compose configuration
├── Dockerfile         # Docker build file
└── go.mod             # Go dependencies
```

## Deployment

### Railway

1. Connect your repository to Railway
2. Add environment variables
3. Railway will automatically detect and deploy

### Render

1. Create a new Web Service
2. Connect your repository
3. Set build command: `go build -o main`
4. Set start command: `./main`
5. Add PostgreSQL database service
6. Configure environment variables

### EC2

1. Set up an EC2 instance
2. Install Docker and Docker Compose
3. Clone the repository
4. Run `docker-compose up -d`
5. Configure security groups for ports 8080 and 2358

## Environment Variables

- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: codesprint)
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name (default: codesprint)
- `JWT_SECRET` - Secret key for JWT tokens (change in production!)
- `JUDGE0_URL` - Judge0 API URL (default: http://localhost:2358)
- `PORT` - Server port (default: 8080)

## Troubleshooting

### Judge0 not responding
- Wait a few minutes for Judge0 to fully initialize
- Check Judge0 logs: `docker-compose logs judge0`
- Verify Judge0 is accessible: `curl http://localhost:2358/status`

### Database connection errors
- Ensure PostgreSQL is running: `docker-compose ps`
- Check database logs: `docker-compose logs postgres`
- Verify environment variables are set correctly

### Frontend not loading
- Ensure the frontend directory exists and contains files
- Check browser console for errors
- Verify the backend is serving static files correctly

## Future Enhancements (Post-MVP)

- WebSocket support for real-time leaderboard updates
- Support for more programming languages
- Plagiarism detection
- Advanced penalty rules
- Payment integration
- Enhanced UI/UX
- Admin dashboard
- Submission history and statistics

## License

MIT License

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.


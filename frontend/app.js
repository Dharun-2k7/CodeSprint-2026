// API Configuration
const API_BASE = '/api';
let authToken = localStorage.getItem('authToken');
let currentUser = null;
let currentContestId = null;
let leaderboardInterval = null;

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    checkAuth();
    loadContests();
    setupEventListeners();
});

// Authentication
function checkAuth() {
    if (authToken) {
        // Token exists, show user info
        const userData = localStorage.getItem('userData');
        if (userData) {
            currentUser = JSON.parse(userData);
            updateUIForAuth();
        }
    }
}

function updateUIForAuth() {
    document.getElementById('login-btn').style.display = 'none';
    document.getElementById('signup-btn').style.display = 'none';
    document.getElementById('logout-btn').style.display = 'block';
    document.getElementById('admin-btn').style.display = 'block';
    document.getElementById('user-info').style.display = 'block';
    document.getElementById('user-info').textContent = `Welcome, ${currentUser.name}`;
}

function setupEventListeners() {
    document.getElementById('login-btn').addEventListener('click', () => showView('login'));
    document.getElementById('signup-btn').addEventListener('click', () => showView('signup'));
    document.getElementById('logout-btn').addEventListener('click', logout);
    document.getElementById('admin-btn').addEventListener('click', () => showView('admin'));
    document.getElementById('login-form').addEventListener('submit', handleLogin);
    document.getElementById('signup-form').addEventListener('submit', handleSignup);
    document.getElementById('create-contest-form').addEventListener('submit', handleCreateContest);
    document.getElementById('create-problem-form').addEventListener('submit', handleCreateProblem);
    document.getElementById('create-testcase-form').addEventListener('submit', handleCreateTestcase);
}

function showView(viewName) {
    document.getElementById('home-view').style.display = viewName === 'home' ? 'block' : 'none';
    document.getElementById('login-view').style.display = viewName === 'login' ? 'block' : 'none';
    document.getElementById('signup-view').style.display = viewName === 'signup' ? 'block' : 'none';
    document.getElementById('contest-view').style.display = viewName === 'contest' ? 'block' : 'none';
}

async function handleLogin(e) {
    e.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        const response = await fetch(`${API_BASE}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await response.json();
        if (response.ok) {
            authToken = data.token;
            currentUser = data.user;
            localStorage.setItem('authToken', authToken);
            localStorage.setItem('userData', JSON.stringify(currentUser));
            updateUIForAuth();
            showView('home');
            loadContests();
        } else {
            alert(data.error || 'Login failed');
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

async function handleSignup(e) {
    e.preventDefault();
    const name = document.getElementById('signup-name').value;
    const email = document.getElementById('signup-email').value;
    const password = document.getElementById('signup-password').value;

    try {
        const response = await fetch(`${API_BASE}/signup`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, email, password })
        });

        const data = await response.json();
        if (response.ok) {
            authToken = data.token;
            currentUser = data.user;
            localStorage.setItem('authToken', authToken);
            localStorage.setItem('userData', JSON.stringify(currentUser));
            updateUIForAuth();
            showView('home');
            loadContests();
        } else {
            alert(data.error || 'Signup failed');
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

function logout() {
    authToken = null;
    currentUser = null;
    localStorage.removeItem('authToken');
    localStorage.removeItem('userData');
    document.getElementById('login-btn').style.display = 'block';
    document.getElementById('signup-btn').style.display = 'block';
    document.getElementById('logout-btn').style.display = 'none';
    document.getElementById('user-info').style.display = 'none';
    showView('home');
    if (leaderboardInterval) {
        clearInterval(leaderboardInterval);
    }
}

// Contests
async function loadContests() {
    try {
        const response = await fetch(`${API_BASE}/contests`);
        const contests = await response.json();
        displayContests(contests);
    } catch (error) {
        console.error('Error loading contests:', error);
    }
}

function displayContests(contests) {
    const container = document.getElementById('contests-list');
    container.innerHTML = '<h3>Available Contests</h3>';
    
    if (contests.length === 0) {
        container.innerHTML += '<p>No contests available</p>';
        return;
    }

    contests.forEach(contest => {
        const card = document.createElement('div');
        card.className = 'card contest-card';
        card.innerHTML = `
            <div class="card-body">
                <h5 class="card-title">${contest.title}</h5>
                <p class="card-text">
                    Start: ${new Date(contest.start_time).toLocaleString()}<br>
                    End: ${new Date(contest.end_time).toLocaleString()}
                </p>
            </div>
        `;
        card.addEventListener('click', () => openContest(contest.id));
        container.appendChild(card);
    });
}

async function openContest(contestId) {
    if (!authToken) {
        alert('Please login to view contests');
        showView('login');
        return;
    }

    currentContestId = contestId;
    showView('contest');
    
    try {
        const [contest, problems] = await Promise.all([
            fetch(`${API_BASE}/contest/${contestId}`).then(r => r.json()),
            fetch(`${API_BASE}/problems?contest_id=${contestId}`, {
                headers: { 'Authorization': `Bearer ${authToken}` }
            }).then(r => r.json())
        ]);

        document.getElementById('contest-title').textContent = contest.title;
        displayProblems(problems);
        loadLeaderboard(contestId);
        
        // Auto-refresh leaderboard every 10 seconds
        if (leaderboardInterval) {
            clearInterval(leaderboardInterval);
        }
        leaderboardInterval = setInterval(() => loadLeaderboard(contestId), 10000);
    } catch (error) {
        console.error('Error loading contest:', error);
    }
}

function displayProblems(problems) {
    const container = document.getElementById('problems-list');
    container.innerHTML = '<h4>Problems</h4>';
    
    problems.forEach(problem => {
        const card = document.createElement('div');
        card.className = 'card problem-card';
        card.innerHTML = `
            <div class="card-body">
                <h5 class="card-title">${problem.title}</h5>
                <p class="card-text">Time Limit: ${problem.time_limit}ms | Memory Limit: ${problem.memory_limit}MB</p>
                <button class="btn btn-primary" onclick="viewProblem(${problem.id})">View Problem</button>
            </div>
        `;
        container.appendChild(card);
    });
}

async function viewProblem(problemId) {
    try {
        const response = await fetch(`${API_BASE}/problem/${problemId}`);
        const problem = await response.json();
        
        const detailDiv = document.getElementById('problem-detail');
        detailDiv.innerHTML = `
            <h4>${problem.title}</h4>
            <div class="problem-statement">${problem.statement}</div>
            <button class="btn btn-success" onclick="showSubmissionForm(${problem.id})">Submit Solution</button>
        `;
        detailDiv.style.display = 'block';
        
        // Load sample testcases
        const testcasesResponse = await fetch(`${API_BASE}/testcases?problem_id=${problemId}`);
        const testcases = await testcasesResponse.json();
        if (testcases.length > 0) {
            let testcasesHtml = '<h5>Sample Test Cases</h5>';
            testcases.forEach((tc, idx) => {
                testcasesHtml += `
                    <div class="card mb-2">
                        <div class="card-body">
                            <strong>Input ${idx + 1}:</strong>
                            <pre>${tc.input}</pre>
                            <strong>Expected Output ${idx + 1}:</strong>
                            <pre>${tc.expected_output}</pre>
                        </div>
                    </div>
                `;
            });
            detailDiv.innerHTML += testcasesHtml;
        }
    } catch (error) {
        console.error('Error loading problem:', error);
    }
}

function showSubmissionForm(problemId) {
    const formDiv = document.getElementById('submission-form');
    formDiv.innerHTML = `
        <h4>Submit Solution</h4>
        <form id="code-submission-form">
            <div class="mb-3">
                <label for="language" class="form-label">Language</label>
                <select class="form-select" id="language" required>
                    <option value="c">C</option>
                    <option value="cpp">C++</option>
                    <option value="python3">Python 3</option>
                </select>
            </div>
            <div class="mb-3">
                <label for="code" class="form-label">Code</label>
                <textarea class="form-control code-editor" id="code" required></textarea>
            </div>
            <button type="submit" class="btn btn-primary">Submit</button>
        </form>
    `;
    formDiv.style.display = 'block';
    
    document.getElementById('code-submission-form').addEventListener('submit', (e) => {
        e.preventDefault();
        submitCode(problemId);
    });
}

async function submitCode(problemId) {
    const language = document.getElementById('language').value;
    const code = document.getElementById('code').value;

    try {
        const response = await fetch(`${API_BASE}/submission`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({
                problem_id: problemId,
                contest_id: currentContestId,
                language: language,
                code: code
            })
        });

        const data = await response.json();
        if (response.ok) {
            alert(`Submission received! ID: ${data.submission_id}. Status: ${data.status}`);
            // Poll for result
            pollSubmissionResult(data.submission_id);
        } else {
            alert(data.error || 'Submission failed');
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

async function pollSubmissionResult(submissionId) {
    const maxAttempts = 30;
    let attempts = 0;
    
    const poll = async () => {
        try {
            const response = await fetch(`${API_BASE}/submission/${submissionId}`);
            const submission = await response.json();
            
            if (submission.status !== 'pending' && submission.status !== 'running') {
                alert(`Submission ${submissionId}: ${submission.status.toUpperCase()}\nScore: ${submission.score}\nRuntime: ${submission.runtime}ms`);
                if (currentContestId) {
                    loadLeaderboard(currentContestId);
                }
                return;
            }
            
            attempts++;
            if (attempts < maxAttempts) {
                setTimeout(poll, 2000);
            } else {
                alert('Submission is taking longer than expected. Please check later.');
            }
        } catch (error) {
            console.error('Error polling submission:', error);
        }
    };
    
    poll();
}

async function loadLeaderboard(contestId) {
    try {
        const response = await fetch(`${API_BASE}/leaderboard/${contestId}`);
        const leaderboard = await response.json();
        displayLeaderboard(leaderboard);
    } catch (error) {
        console.error('Error loading leaderboard:', error);
    }
}

function displayLeaderboard(leaderboard) {
    const container = document.getElementById('leaderboard');
    container.innerHTML = '<h4>Leaderboard</h4><div class="leaderboard">';
    
    if (leaderboard.length === 0) {
        container.innerHTML += '<p>No submissions yet</p></div>';
        return;
    }

    leaderboard.forEach((entry, index) => {
        const entryDiv = document.createElement('div');
        entryDiv.className = 'leaderboard-entry';
        entryDiv.innerHTML = `
            <strong>#${index + 1}</strong> ${entry.user_name}<br>
            Solved: ${entry.solved_count} | Penalty: ${entry.penalty} minutes
        `;
        container.querySelector('.leaderboard').appendChild(entryDiv);
    });
    
    container.innerHTML += '</div>';
}

// Make functions available globally
window.viewProblem = viewProblem;
window.showSubmissionForm = showSubmissionForm;


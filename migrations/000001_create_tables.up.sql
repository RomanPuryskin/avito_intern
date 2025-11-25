BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(255) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(100) PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    team_name VARCHAR(255) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);


CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(100) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(100) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    "status" VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_pr_author_id ON pull_requests(author_id);

CREATE TABLE IF NOT EXISTS assigned_reviewers (
    pull_request_id VARCHAR(100) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id VARCHAR(100) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (pull_request_id, user_id)
);
CREATE INDEX idx_assigned_reviewers_pr_id ON assigned_reviewers(pull_request_id);
CREATE INDEX idx_assigned_reviewers_user_id ON assigned_reviewers(user_id);

COMMIT;
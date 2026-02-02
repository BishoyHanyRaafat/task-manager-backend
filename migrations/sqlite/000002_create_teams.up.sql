CREATE TABLE IF NOT EXISTS teams
(
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Many-to-many relation
CREATE TABLE IF NOT EXISTS teams_users
(
    id      TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role    TEXT NOT NULL CHECK (role IN ('founder', 'admin', 'standard')) DEFAULT 'standard',
    FOREIGN KEY (team_id) REFERENCES teams (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS tasks
(
    id           TEXT PRIMARY KEY,
    team_id      TEXT    NOT NULL,
    grouped      BOOLEAN NOT NULL, -- Individual or group task
    user_id_task TEXT    NOT NULL,
    FOREIGN KEY (user_id_task) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (team_id) REFERENCES teams (id) ON DELETE CASCADE
)
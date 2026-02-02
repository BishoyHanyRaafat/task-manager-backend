CREATE TYPE TEAM_ROLE AS ENUM ('founder', 'admin', 'standard');

CREATE TABLE IF NOT EXISTS teams
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Many-to-many relation
CREATE TABLE IF NOT EXISTS teams_users
(
    id      UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
    team_id UUID      NOT NULL,
    user_id UUID      NOT NULL,
    role    TEAM_ROLE NOT NULL DEFAULT 'standard',
    FOREIGN KEY (team_id) REFERENCES teams (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS tasks
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id      UUID    NOT NULL,
    grouped      BOOLEAN NOT NULL, -- Individual or group task
    user_id_task UUID,             -- NOT NULL If individual
    FOREIGN KEY (user_id_task) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (team_id) REFERENCES teams (id) ON DELETE CASCADE
)
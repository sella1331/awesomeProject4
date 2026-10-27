CREATE TABLE tasks (
                       task_id      UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
                       user_id      UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                       title        VARCHAR(255) NOT NULL,
                       description  TEXT         NOT NULL,
                       date_created TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                       urgency      VARCHAR(50)  NOT NULL
                           CHECK (urgency IN ('low','medium','high')),
                       status       VARCHAR(50)  NOT NULL DEFAULT 'new'
                           CHECK (status IN ('new','in_progress','finish'))
);CREATE UNIQUE INDEX idx_tasks_user_title_desc
    ON tasks (user_id, title, description);
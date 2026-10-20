CREATE TABLE IF NOT EXISTS ai_model_usage (
    name TEXT PRIMARY KEY NOT NULL,
    usedRequestsDay INTEGER NOT NULL DEFAULT 0,
    usedRequestsMin INTEGER NOT NULL DEFAULT 0,
    nextResetDay DATETIME NOT NULL,
    nextResetMin DATETIME NOT NULL
);

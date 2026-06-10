CREATE TABLE IF NOT EXISTS requests (
  id                    TEXT    PRIMARY KEY,
  timestamp             INTEGER NOT NULL,
  day_date              TEXT    NOT NULL,
  provider              TEXT    NOT NULL,
  model                 TEXT    NOT NULL,
  task_type             TEXT,
  complexity            TEXT,
  tokens_in             INTEGER,
  tokens_out            INTEGER,
  cost_usd              REAL,
  latency_ms            INTEGER,
  content_mode          INTEGER NOT NULL DEFAULT 0,
  classifier_source     TEXT,
  classifier_confidence REAL,
  tool_source           TEXT,
  intercepted           INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_req_day   ON requests(day_date);
CREATE INDEX IF NOT EXISTS idx_req_model ON requests(model);
CREATE INDEX IF NOT EXISTS idx_req_task  ON requests(task_type, model);

CREATE TABLE IF NOT EXISTS daily_reports (
  date                  TEXT    PRIMARY KEY,
  generated_at          INTEGER NOT NULL,
  total_requests        INTEGER NOT NULL DEFAULT 0,
  total_tokens_in       INTEGER NOT NULL DEFAULT 0,
  total_tokens_out      INTEGER NOT NULL DEFAULT 0,
  total_cost_usd        REAL    NOT NULL DEFAULT 0.0,
  optimized_cost_usd    REAL    NOT NULL DEFAULT 0.0,
  savings_potential_usd REAL    NOT NULL DEFAULT 0.0,
  report_json           TEXT
);

CREATE TABLE IF NOT EXISTS config (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

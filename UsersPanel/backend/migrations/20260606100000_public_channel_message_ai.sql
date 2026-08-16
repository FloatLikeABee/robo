-- Public channel + AI digest metadata for platform messaging

ALTER TABLE msg_threads
    ADD COLUMN is_public_channel TINYINT(1) NOT NULL DEFAULT 0;

ALTER TABLE msg_messages
    ADD COLUMN message_kind VARCHAR(32) NOT NULL DEFAULT 'human';

CREATE TABLE IF NOT EXISTS msg_ai_digest_state (
    id TINYINT NOT NULL PRIMARY KEY DEFAULT 1,
    last_run_at VARCHAR(64) NULL,
    last_events_at VARCHAR(64) NULL,
    last_docs_sync_at VARCHAR(64) NULL,
    digests_today INT NOT NULL DEFAULT 0,
    digest_day VARCHAR(10) NULL,
    updated_at VARCHAR(64) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO msg_ai_digest_state (id, updated_at)
VALUES (1, '2026-06-06T00:00:00Z');

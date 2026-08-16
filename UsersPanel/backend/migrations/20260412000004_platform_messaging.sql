-- Platform in-system messaging (direct + group threads)

CREATE TABLE IF NOT EXISTS msg_threads (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    title VARCHAR(255) NULL,
    created_by VARCHAR(36) NOT NULL,
    is_group TINYINT(1) NOT NULL DEFAULT 0,
    created_at VARCHAR(64) NOT NULL,
    updated_at VARCHAR(64) NOT NULL,
    KEY idx_msg_threads_updated_at (updated_at),
    CONSTRAINT fk_msg_threads_created_by FOREIGN KEY (created_by)
        REFERENCES plat_users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS msg_thread_members (
    thread_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    joined_at VARCHAR(64) NOT NULL,
    PRIMARY KEY (thread_id, user_id),
    KEY idx_msg_thread_members_user_id (user_id),
    CONSTRAINT fk_msg_thread_members_thread FOREIGN KEY (thread_id)
        REFERENCES msg_threads (id) ON DELETE CASCADE,
    CONSTRAINT fk_msg_thread_members_user FOREIGN KEY (user_id)
        REFERENCES plat_users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS msg_messages (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    thread_id VARCHAR(36) NOT NULL,
    sender_user_id VARCHAR(36) NOT NULL,
    body TEXT NOT NULL,
    created_at VARCHAR(64) NOT NULL,
    KEY idx_msg_messages_thread_created (thread_id, created_at),
    CONSTRAINT fk_msg_messages_thread FOREIGN KEY (thread_id)
        REFERENCES msg_threads (id) ON DELETE CASCADE,
    CONSTRAINT fk_msg_messages_sender FOREIGN KEY (sender_user_id)
        REFERENCES plat_users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS msg_message_reads (
    message_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    read_at VARCHAR(64) NOT NULL,
    PRIMARY KEY (message_id, user_id),
    KEY idx_msg_message_reads_user_id (user_id),
    CONSTRAINT fk_msg_reads_message FOREIGN KEY (message_id)
        REFERENCES msg_messages (id) ON DELETE CASCADE,
    CONSTRAINT fk_msg_reads_user FOREIGN KEY (user_id)
        REFERENCES plat_users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Fast unread scans per user.
CREATE TABLE IF NOT EXISTS msg_deliveries (
    message_id VARCHAR(36) NOT NULL,
    recipient_user_id VARCHAR(36) NOT NULL,
    delivered_at VARCHAR(64) NOT NULL,
    PRIMARY KEY (message_id, recipient_user_id),
    KEY idx_msg_deliveries_recipient_delivered (recipient_user_id, delivered_at),
    CONSTRAINT fk_msg_deliveries_message FOREIGN KEY (message_id)
        REFERENCES msg_messages (id) ON DELETE CASCADE,
    CONSTRAINT fk_msg_deliveries_user FOREIGN KEY (recipient_user_id)
        REFERENCES plat_users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO plat_permissions (id, name, description, created_at, updated_at) VALUES
('11111111-1111-4111-8111-111111111110', 'inbox_message', 'Send and read in-system messages', '2026-05-14T00:00:00Z', '2026-05-14T00:00:00Z');

INSERT IGNORE INTO plat_role_permissions (role_name, permission_id) VALUES
('Main Panel', '11111111-1111-4111-8111-111111111110'),
('Forms', '11111111-1111-4111-8111-111111111110'),
('Email Composer', '11111111-1111-4111-8111-111111111110'),
('Sharp Reports', '11111111-1111-4111-8111-111111111110'),
('Admin', '11111111-1111-4111-8111-111111111110');

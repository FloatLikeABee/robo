-- Cases & Tasks module

CREATE TABLE IF NOT EXISTS CaseTask (
  ID INT NOT NULL AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  description TEXT NULL,
  start_at DATETIME NULL,
  end_at DATETIME NULL,
  location JSON NULL,
  assignee_type ENUM('member', 'employee') NOT NULL,
  assignee_id INT NOT NULL,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  KEY ix_case_task_assignee (assignee_type, assignee_id),
  KEY ix_case_task_start_at (start_at)
);

CREATE TABLE IF NOT EXISTS CaseTaskAttachment (
  id INT NOT NULL AUTO_INCREMENT,
  case_task_id INT NOT NULL,
  original_name VARCHAR(255) NOT NULL,
  stored_name VARCHAR(255) NOT NULL,
  file_path VARCHAR(600) NOT NULL,
  mime_type VARCHAR(120) NULL,
  size_bytes BIGINT NOT NULL DEFAULT 0,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY ix_case_task_attachment_case_task_id (case_task_id),
  CONSTRAINT fk_case_task_attachment_case_task FOREIGN KEY (case_task_id) REFERENCES CaseTask (ID) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Seed a few mock rows if table is empty.
INSERT INTO CaseTask (title, description, start_at, end_at, location, assignee_type, assignee_id)
SELECT *
FROM (
  SELECT
    'Route delay follow-up' AS title,
    'Investigate recurring AM route delay and assign mitigations.' AS description,
    DATE_ADD(NOW(), INTERVAL 1 DAY) AS start_at,
    DATE_ADD(DATE_ADD(NOW(), INTERVAL 1 DAY), INTERVAL 90 MINUTE) AS end_at,
    JSON_OBJECT(
      'label', 'East Valley corridor',
      'area', JSON_ARRAY(
        JSON_ARRAY(34.142, -118.121),
        JSON_ARRAY(34.137, -118.101),
        JSON_ARRAY(34.121, -118.112),
        JSON_ARRAY(34.126, -118.131)
      )
    ) AS location,
    'employee' AS assignee_type,
    e.id AS assignee_id
  FROM employee e
  ORDER BY e.id
  LIMIT 1
) seed1
WHERE (SELECT COUNT(*) FROM CaseTask) = 0
UNION ALL
SELECT *
FROM (
  SELECT
    'Member assistance request' AS title,
    'Case intake for transportation assistance change request.' AS description,
    DATE_ADD(NOW(), INTERVAL 2 DAY) AS start_at,
    DATE_ADD(DATE_ADD(NOW(), INTERVAL 2 DAY), INTERVAL 45 MINUTE) AS end_at,
    NULL AS location,
    'member' AS assignee_type,
    m.id AS assignee_id
  FROM `member` m
  ORDER BY m.id
  LIMIT 1
) seed2
WHERE (SELECT COUNT(*) FROM CaseTask) = 0
UNION ALL
SELECT *
FROM (
  SELECT
    'Fleet document review' AS title,
    'Review maintenance docs and close outstanding validation task.' AS description,
    DATE_ADD(NOW(), INTERVAL 3 DAY) AS start_at,
    DATE_ADD(DATE_ADD(NOW(), INTERVAL 3 DAY), INTERVAL 60 MINUTE) AS end_at,
    JSON_OBJECT('label', 'Main depot area') AS location,
    'employee' AS assignee_type,
    e.id AS assignee_id
  FROM employee e
  ORDER BY e.id DESC
  LIMIT 1
) seed3
WHERE (SELECT COUNT(*) FROM CaseTask) = 0;

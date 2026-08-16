-- CaseTask: multiple assignees (member | employee | contact) via junction table.

CREATE TABLE IF NOT EXISTS CaseTaskAssignee (
  id INT NOT NULL AUTO_INCREMENT,
  case_task_id INT NOT NULL,
  assignee_kind ENUM('member', 'employee', 'contact') NOT NULL,
  assignee_id INT NOT NULL,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY ux_case_task_assignee (case_task_id, assignee_kind, assignee_id),
  KEY ix_case_task_assignee_task (case_task_id),
  CONSTRAINT fk_case_task_assignee_task FOREIGN KEY (case_task_id) REFERENCES CaseTask (ID) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Backfill from legacy columns (idempotent).
INSERT IGNORE INTO CaseTaskAssignee (case_task_id, assignee_kind, assignee_id)
SELECT c.ID, c.assignee_type, c.assignee_id
FROM CaseTask c
WHERE c.assignee_id IS NOT NULL AND c.assignee_id > 0 AND c.assignee_type IS NOT NULL AND TRIM(c.assignee_type) != '';

-- Allow NULL legacy columns when assignees live in junction only.
SET @db := DATABASE();
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'CaseTask' AND column_name = 'assignee_type' AND is_nullable = 'NO');
SET @q := IF(@c > 0, 'ALTER TABLE CaseTask MODIFY assignee_type ENUM(''member'',''employee'') NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'CaseTask' AND column_name = 'assignee_id' AND is_nullable = 'NO');
SET @q := IF(@c > 0, 'ALTER TABLE CaseTask MODIFY assignee_id INT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

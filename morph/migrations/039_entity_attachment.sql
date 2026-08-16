-- Unified file attachments for admin entities (facility, member, employee, asset, activity, case_task).

CREATE TABLE IF NOT EXISTS EntityAttachment (
  id INT NOT NULL AUTO_INCREMENT,
  entity_type VARCHAR(32) NOT NULL,
  record_id INT NOT NULL,
  original_name VARCHAR(255) NOT NULL,
  stored_name VARCHAR(255) NOT NULL,
  file_path VARCHAR(600) NOT NULL,
  mime_type VARCHAR(120) NULL,
  size_bytes BIGINT NOT NULL DEFAULT 0,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY ix_entity_attachment_entity (entity_type, record_id)
);

-- Migrate legacy case/task attachments when present.
INSERT INTO EntityAttachment (entity_type, record_id, original_name, stored_name, file_path, mime_type, size_bytes, created_on)
SELECT 'case_task', case_task_id, original_name, stored_name, file_path, mime_type, size_bytes, created_on
FROM CaseTaskAttachment
WHERE NOT EXISTS (
  SELECT 1 FROM EntityAttachment ea
  WHERE ea.entity_type = 'case_task' AND ea.record_id = CaseTaskAttachment.case_task_id
    AND ea.stored_name = CaseTaskAttachment.stored_name
);

DROP TABLE IF EXISTS CaseTaskAttachment;

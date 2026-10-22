-- Per-user message AI auto-reply settings + dedupe log for processed inbound messages.

SET @dbname = DATABASE();

SET @col_enabled = (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = 'User' AND COLUMN_NAME = 'MessageAiAutoReplyEnabled'
);
SET @sql_enabled = IF(@col_enabled = 0,
  'ALTER TABLE `User` ADD COLUMN `MessageAiAutoReplyEnabled` TINYINT(1) NOT NULL DEFAULT 0 AFTER `Title`',
  'SELECT 1');
PREPARE m038_enabled FROM @sql_enabled;
EXECUTE m038_enabled;
DEALLOCATE PREPARE m038_enabled;

SET @col_prompt = (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = 'User' AND COLUMN_NAME = 'MessageAiAutoReplyPrompt'
);
SET @sql_prompt = IF(@col_prompt = 0,
  'ALTER TABLE `User` ADD COLUMN `MessageAiAutoReplyPrompt` LONGTEXT NULL AFTER `MessageAiAutoReplyEnabled`',
  'SELECT 1');
PREPARE m038_prompt FROM @sql_prompt;
EXECUTE m038_prompt;
DEALLOCATE PREPARE m038_prompt;

CREATE TABLE IF NOT EXISTS `user_message_auto_reply_log` (
  ID INT NOT NULL AUTO_INCREMENT,
  UserID INT NOT NULL,
  ThreadID VARCHAR(64) NOT NULL,
  MessageID VARCHAR(64) NOT NULL,
  CreatedOn DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  UNIQUE KEY UX_user_message_auto_reply_log (UserID, MessageID),
  KEY IX_user_message_auto_reply_log_user (UserID, CreatedOn)
);

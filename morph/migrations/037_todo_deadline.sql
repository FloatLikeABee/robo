-- TODO deadline for schedules (idempotent: safe if column / index already exist).

SET @dbname = DATABASE();

SET @col_exists = (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = 'user_note_todo' AND COLUMN_NAME = 'DeadlineAt'
);

SET @sqladdcol = IF(@col_exists = 0,
  'ALTER TABLE `user_note_todo` ADD COLUMN `DeadlineAt` DATETIME NULL AFTER `Completed`',
  'SELECT 1');
PREPARE m037_col FROM @sqladdcol;
EXECUTE m037_col;
DEALLOCATE PREPARE m037_col;

SET @idx_exists = (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = 'user_note_todo' AND INDEX_NAME = 'IX_user_note_todo_deadline'
);

SET @sqladdidx = IF(@idx_exists = 0,
  'ALTER TABLE `user_note_todo` ADD KEY `IX_user_note_todo_deadline` (`UserID`, `ItemType`, `DeadlineAt`)',
  'SELECT 1');
PREPARE m037_idx FROM @sqladdidx;
EXECUTE m037_idx;
DEALLOCATE PREPARE m037_idx;

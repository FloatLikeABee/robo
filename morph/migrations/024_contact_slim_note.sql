-- contact: drop legacy columns; Comments -> Note.
-- MySQL 5.7+ / MariaDB.

SET @db := DATABASE();
SET FOREIGN_KEY_CHECKS = 0;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'DBID');
SET @q := IF(@c > 0, 'ALTER TABLE `contact` DROP COLUMN DBID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'Mail_Street1');
SET @q := IF(@c > 0, 'ALTER TABLE `contact` DROP COLUMN Mail_Street1', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'Mail_Street2');
SET @q := IF(@c > 0, 'ALTER TABLE `contact` DROP COLUMN Mail_Street2', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'Mail_City');
SET @q := IF(@c > 0, 'ALTER TABLE `contact` DROP COLUMN Mail_City', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'Mail_Zip');
SET @q := IF(@c > 0, 'ALTER TABLE `contact` DROP COLUMN Mail_Zip', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'Mail_State_Id');
SET @q := IF(@c > 0, 'ALTER TABLE `contact` DROP COLUMN Mail_State_Id', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'CreatedOn');
SET @q := IF(@c > 0, 'ALTER TABLE `contact` DROP COLUMN CreatedOn', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'LastUpdated');
SET @q := IF(@c > 0, 'ALTER TABLE `contact` DROP COLUMN LastUpdated', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'Comments');
SET @q := IF(@c > 0, 'ALTER TABLE `contact` CHANGE COLUMN Comments Note LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET FOREIGN_KEY_CHECKS = 1;

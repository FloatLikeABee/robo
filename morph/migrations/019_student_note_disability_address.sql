-- Student: drop legacy columns; Comments -> Note; Disabled -> DisabilityCodeId (FK); address via CityId, PostalCodeId, StateId.
-- Compatible with MySQL 5.7+ / MariaDB (no DROP INDEX IF EXISTS / DROP COLUMN IF EXISTS — those need MySQL 8.0.29+).

SET @db := DATABASE();
SET FOREIGN_KEY_CHECKS = 0;

-- Drop legacy indexes (only if present)
SET @ix1 := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Student' AND index_name = 'UIX_Student_DBID_StudID'
);
SET @s1 := IF(@ix1 > 0, 'ALTER TABLE Student DROP INDEX UIX_Student_DBID_StudID', 'SELECT 1');
PREPARE ps1 FROM @s1;
EXECUTE ps1;
DEALLOCATE PREPARE ps1;

SET @ix2 := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Student' AND index_name = 'IX_Student_DBID_School'
);
SET @s2 := IF(@ix2 > 0, 'ALTER TABLE Student DROP INDEX IX_Student_DBID_School', 'SELECT 1');
PREPARE ps2 FROM @s2;
EXECUTE ps2;
DEALLOCATE PREPARE ps2;

-- Drop legacy columns (one at a time; only if column exists)
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'DBID');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN DBID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Stud_ID');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN Stud_ID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'DistrictID');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN DistrictID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Mail_Street1');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN Mail_Street1', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Mail_Street2');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN Mail_Street2', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Mail_City');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN Mail_City', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Mail_Zip');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN Mail_Zip', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'DistanceFromResidSch');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN DistanceFromResidSch', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'PopulationRegionID');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN PopulationRegionID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Mail_State_Id');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN Mail_State_Id', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'CreatedOn');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN CreatedOn', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'CreatedBy');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN CreatedBy', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Comments -> Note
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Comments');
SET @q := IF(@c > 0, 'ALTER TABLE Student CHANGE COLUMN Comments Note LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Drop Disabled
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Disabled');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN Disabled', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- New columns (skip if already migrated)
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'DisabilityCodeId');
SET @q := IF(@c = 0, 'ALTER TABLE Student ADD COLUMN DisabilityCodeId INT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'CityId');
SET @q := IF(@c = 0, 'ALTER TABLE Student ADD COLUMN CityId INT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'PostalCodeId');
SET @q := IF(@c = 0, 'ALTER TABLE Student ADD COLUMN PostalCodeId INT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'StateId');
SET @q := IF(@c = 0, 'ALTER TABLE Student ADD COLUMN StateId INT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Foreign keys (only if not already present)
SET @fk := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Student' AND constraint_name = 'FK_Student_DisabilityCode'
);
SET @q := IF(@fk = 0, 'ALTER TABLE Student ADD CONSTRAINT FK_Student_DisabilityCode FOREIGN KEY (DisabilityCodeId) REFERENCES Disability_Codes (ID) ON DELETE SET NULL ON UPDATE CASCADE', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @fk := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Student' AND constraint_name = 'FK_Student_MailingCity'
);
SET @q := IF(@fk = 0, 'ALTER TABLE Student ADD CONSTRAINT FK_Student_MailingCity FOREIGN KEY (CityId) REFERENCES MailingCity (ID) ON DELETE SET NULL ON UPDATE CASCADE', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @fk := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Student' AND constraint_name = 'FK_Student_MailingPostalCode'
);
SET @q := IF(@fk = 0, 'ALTER TABLE Student ADD CONSTRAINT FK_Student_MailingPostalCode FOREIGN KEY (PostalCodeId) REFERENCES MailingPostalCode (ID) ON DELETE SET NULL ON UPDATE CASCADE', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @fk := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Student' AND constraint_name = 'FK_Student_MailingState'
);
SET @q := IF(@fk = 0, 'ALTER TABLE Student ADD CONSTRAINT FK_Student_MailingState FOREIGN KEY (StateId) REFERENCES MailingState (ID) ON DELETE SET NULL ON UPDATE CASCADE', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- School index (if missing)
SET @ix_school := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Student' AND index_name = 'IX_Student_School'
);
SET @sql_ix := IF(@ix_school = 0, 'ALTER TABLE Student ADD INDEX IX_Student_School (School)', 'SELECT 1');
PREPARE stmt_ix FROM @sql_ix;
EXECUTE stmt_ix;
DEALLOCATE PREPARE stmt_ix;

SET FOREIGN_KEY_CHECKS = 1;

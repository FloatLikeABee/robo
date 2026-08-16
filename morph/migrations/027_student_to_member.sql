-- Student table -> `member` (MySQL: MEMBER is reserved; backtick the table name).
-- Drop DisabilityCodeId, Grade; School -> facility; all columns snake_case; StudentSchedule.StudentID -> member_id.
SET @db := DATABASE();
SET @had_student := (
  SELECT COUNT(*) FROM information_schema.tables
  WHERE table_schema = @db AND table_name = 'Student'
);

-- FK: DisabilityCodeId
SET @fk := (SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Student' AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'FK_Student_DisabilityCode');
SET @q := IF(@fk > 0 AND @had_student > 0, 'ALTER TABLE Student DROP FOREIGN KEY FK_Student_DisabilityCode', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'DisabilityCodeId');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN DisabilityCodeId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Grade');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN Grade', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Index on old School column
SET @ix := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Student' AND index_name = 'IX_Student_School');
SET @q := IF(@ix > 0, 'ALTER TABLE Student DROP INDEX IX_Student_School', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'School');
SET @q := IF(@c > 0, "ALTER TABLE Student CHANGE COLUMN School facility VARCHAR(10) NULL", 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Rename remaining columns to snake_case
SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'ID');
SET @q := IF(@c > 0, 'ALTER TABLE Student CHANGE COLUMN ID id INT NOT NULL AUTO_INCREMENT', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Last_Name');
SET @q := IF(@c > 0, 'ALTER TABLE Student CHANGE COLUMN Last_Name last_name VARCHAR(20) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'First_Name');
SET @q := IF(@c > 0, 'ALTER TABLE Student CHANGE COLUMN First_Name first_name VARCHAR(15) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Middle_Name');
SET @q := IF(@c > 0, 'ALTER TABLE Student CHANGE COLUMN Middle_Name middle_name VARCHAR(50) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Dob');
SET @q := IF(@c > 0, 'ALTER TABLE Student CHANGE COLUMN Dob dob DATE NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Entry_Date');
SET @q := IF(@c > 0, "ALTER TABLE Student CHANGE COLUMN Entry_Date entry_date DATETIME NULL DEFAULT CURRENT_TIMESTAMP", 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Gender');
SET @q := IF(@c > 0, 'ALTER TABLE Student CHANGE COLUMN Gender gender SMALLINT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Email');
SET @q := IF(@c > 0, 'ALTER TABLE Student CHANGE COLUMN Email email VARCHAR(250) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'ParticipantType');
SET @q := IF(@c > 0, "ALTER TABLE Student CHANGE COLUMN ParticipantType participant_type VARCHAR(64) NULL", 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Index on facility (after renames; column may be facility on Student)
SET @ixf := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Student' AND index_name = 'idx_member_facility');
SET @cf := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'facility');
SET @q := IF(@ixf = 0 AND @cf > 0, 'ALTER TABLE Student ADD INDEX idx_member_facility (facility)', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @q := IF(@had_student > 0, 'RENAME TABLE Student TO `member`', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- StudentSchedule: StudentID -> member_id
SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'StudentSchedule' AND column_name = 'StudentID');
SET @ix2 := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'StudentSchedule' AND index_name = 'IX_StudentSchedule_DBID_StudentID');
SET @q := IF(@c > 0 AND @ix2 > 0, 'ALTER TABLE StudentSchedule DROP INDEX IX_StudentSchedule_DBID_StudentID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'StudentSchedule' AND column_name = 'StudentID');
SET @q := IF(@c > 0, 'ALTER TABLE StudentSchedule CHANGE COLUMN StudentID member_id INT NOT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @ix3 := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'StudentSchedule' AND index_name = 'IX_StudentSchedule_DBID_member_id');
SET @q := IF(@ix3 = 0, 'ALTER TABLE StudentSchedule ADD INDEX IX_StudentSchedule_DBID_member_id (DBID, member_id)', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

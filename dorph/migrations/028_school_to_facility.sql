-- School -> `facility`; SchoolCode -> facility_code; drop FeedSchoolCode, Private; snake_case columns.
-- SchoolGrade.SchoolCode -> facility_code.
SET @db := DATABASE();
SET @had_school := (
  SELECT COUNT(*) FROM information_schema.tables
  WHERE table_schema = @db AND table_name = 'School'
);

-- Unique / indexes that reference old column names
SET @ixu := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'School' AND index_name = 'UQ_School_SchoolCode');
SET @q := IF(@ixu > 0 AND @had_school > 0, 'ALTER TABLE School DROP INDEX UQ_School_SchoolCode', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @ixc := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'School' AND index_name = 'IX_School_XCoordYCoord');
SET @q := IF(@ixc > 0 AND @had_school > 0, 'ALTER TABLE School DROP INDEX IX_School_XCoordYCoord', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'FeedSchoolCode');
SET @q := IF(@c > 0, 'ALTER TABLE School DROP COLUMN FeedSchoolCode', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'Private');
SET @q := IF(@c > 0, 'ALTER TABLE School DROP COLUMN Private', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'ID');
SET @q := IF(@c > 0, 'ALTER TABLE School CHANGE COLUMN ID id INT NOT NULL AUTO_INCREMENT', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'SchoolCode');
SET @q := IF(@c > 0, 'ALTER TABLE School CHANGE COLUMN SchoolCode facility_code VARCHAR(10) NOT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'Name');
SET @q := IF(@c > 0, 'ALTER TABLE School CHANGE COLUMN Name name VARCHAR(150) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'DistrictID');
SET @q := IF(@c > 0, 'ALTER TABLE School CHANGE COLUMN DistrictID district_id INT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'XCoord');
SET @q := IF(@c > 0, 'ALTER TABLE School CHANGE COLUMN XCoord x_coord DECIMAL(9,6) NULL DEFAULT 0', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'YCoord');
SET @q := IF(@c > 0, 'ALTER TABLE School CHANGE COLUMN YCoord y_coord DECIMAL(9,6) NULL DEFAULT 0', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'GUID');
SET @q := IF(@c > 0, 'ALTER TABLE School CHANGE COLUMN GUID guid VARCHAR(50) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'Capacity');
SET @q := IF(@c > 0, 'ALTER TABLE School CHANGE COLUMN Capacity capacity INT NULL DEFAULT 0', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'School' AND column_name = 'FacilityType');
SET @q := IF(@c > 0, "ALTER TABLE School CHANGE COLUMN FacilityType facility_type VARCHAR(64) NULL", 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @q := IF(@had_school > 0, 'RENAME TABLE School TO `facility`', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @ixf := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'facility' AND index_name = 'UQ_facility_facility_code');
SET @tbf := (SELECT COUNT(*) FROM information_schema.tables
  WHERE table_schema = @db AND table_name = 'facility');
SET @q := IF(@ixf = 0 AND @tbf > 0, 'ALTER TABLE `facility` ADD UNIQUE KEY UQ_facility_facility_code (facility_code)', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @ix2 := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'facility' AND index_name = 'idx_facility_xcoord_ycoord');
SET @cxc := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'facility' AND column_name = 'x_coord');
SET @cyc := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'facility' AND column_name = 'y_coord');
SET @q := IF(@ix2 = 0 AND @tbf > 0 AND @cxc > 0 AND @cyc > 0, 'ALTER TABLE `facility` ADD INDEX idx_facility_xcoord_ycoord (x_coord, y_coord)', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- SchoolGrade: link column matches facility_code
SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'SchoolGrade' AND column_name = 'SchoolCode');
SET @q := IF(@c > 0, 'ALTER TABLE SchoolGrade CHANGE COLUMN SchoolCode facility_code VARCHAR(10) NOT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

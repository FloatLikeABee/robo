-- Staff -> `employee`; EMail -> email; CellPhone -> phone_number; drop MailCounty, EmployeeID; snake_case columns.
SET @db := DATABASE();
SET @had_staff := (
  SELECT COUNT(*) FROM information_schema.tables
  WHERE table_schema = @db AND table_name = 'Staff'
);

-- Index on ContractorID (will be renamed)
SET @ixc := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Staff' AND index_name = 'IX_Staff_ContractorID');
SET @q := IF(@ixc > 0 AND @had_staff > 0, 'ALTER TABLE Staff DROP INDEX IX_Staff_ContractorID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'MailCounty');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN MailCounty', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'EmployeeID');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN EmployeeID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'ID');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN ID id INT NOT NULL AUTO_INCREMENT', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'LastName');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN LastName last_name VARCHAR(50) NOT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'FirstName');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN FirstName first_name VARCHAR(50) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'MiddleName');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN MiddleName middle_name VARCHAR(50) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'StaffGUID');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN StaffGUID staff_guid VARCHAR(50) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'ActiveFlag');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN ActiveFlag active_flag TINYINT(1) NOT NULL DEFAULT 1', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'InactiveDate');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN InactiveDate inactive_date DATE NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'ContractorID');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN ContractorID contractor_id INT NULL DEFAULT 0', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'EMail');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN EMail email VARCHAR(100) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'CellPhone');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN CellPhone phone_number VARCHAR(30) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'DateOfBirth');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN DateOfBirth date_of_birth DATE NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'Gender');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN Gender gender SMALLINT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'UserID');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN UserID user_id INT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'EmployType');
SET @q := IF(@c > 0, "ALTER TABLE Staff CHANGE COLUMN EmployType employ_type VARCHAR(64) NULL", 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @q := IF(@had_staff > 0, 'RENAME TABLE Staff TO `employee`', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @tbe := (SELECT COUNT(*) FROM information_schema.tables
  WHERE table_schema = @db AND table_name = 'employee');
SET @ix2 := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'employee' AND index_name = 'idx_employee_contractor_id');
SET @q := IF(@ix2 = 0 AND @tbe > 0, 'ALTER TABLE `employee` ADD INDEX idx_employee_contractor_id (contractor_id)', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

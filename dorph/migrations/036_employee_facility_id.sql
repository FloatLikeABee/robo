-- Optional home / primary facility for employees (links to `facility`.id).

SET @db := DATABASE();

SET @c := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'employee' AND column_name = 'facility_id'
);
SET @q := IF(@c = 0,
  'ALTER TABLE `employee` ADD COLUMN facility_id INT NULL AFTER description',
  'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @fk := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'employee' AND constraint_name = 'fk_employee_facility'
);
SET @q := IF(@fk = 0,
  'ALTER TABLE `employee` ADD CONSTRAINT fk_employee_facility FOREIGN KEY (facility_id) REFERENCES `facility` (id) ON DELETE SET NULL ON UPDATE CASCADE',
  'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

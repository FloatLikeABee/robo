-- Trip -> Activity, Vehicle -> Asset.
-- Activity removes TripDays/Distance/DriverID/VehicleID, adds start/end datetime and JSON location.
-- Asset renames VIN->asset_tag, removes Capacity/ModelInfo/GPSID, adds description.
-- Also adds relation tables: ActivityEmployee, ActivityParticipant, ActivityAsset.

SET @db := DATABASE();
SET @OLD_FK := @@FOREIGN_KEY_CHECKS;
SET FOREIGN_KEY_CHECKS = 0;

-- Rename Trip -> Activity when needed.
SET @has_trip := (
  SELECT COUNT(*) FROM information_schema.tables
  WHERE table_schema = @db AND table_name = 'Trip'
);
SET @has_activity := (
  SELECT COUNT(*) FROM information_schema.tables
  WHERE table_schema = @db AND table_name = 'Activity'
);
SET @q := IF(@has_trip > 0 AND @has_activity = 0, 'RENAME TABLE Trip TO Activity', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Rename Vehicle -> Asset when needed.
SET @has_vehicle := (
  SELECT COUNT(*) FROM information_schema.tables
  WHERE table_schema = @db AND table_name = 'Vehicle'
);
SET @has_asset := (
  SELECT COUNT(*) FROM information_schema.tables
  WHERE table_schema = @db AND table_name = 'Asset'
);
SET @q := IF(@has_vehicle > 0 AND @has_asset = 0, 'RENAME TABLE Vehicle TO Asset', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Ensure relation tables exist first so we can preserve old FK-style columns.
CREATE TABLE IF NOT EXISTS ActivityEmployee (
  id INT NOT NULL AUTO_INCREMENT,
  activity_id INT NOT NULL,
  employee_id INT NOT NULL,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY ux_activity_employee (activity_id, employee_id),
  KEY ix_activity_employee_employee (employee_id),
  CONSTRAINT fk_activity_employee_activity FOREIGN KEY (activity_id) REFERENCES Activity (ID) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_activity_employee_employee FOREIGN KEY (employee_id) REFERENCES employee (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS ActivityParticipant (
  id INT NOT NULL AUTO_INCREMENT,
  activity_id INT NOT NULL,
  member_id INT NOT NULL,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY ux_activity_participant (activity_id, member_id),
  KEY ix_activity_participant_member (member_id),
  CONSTRAINT fk_activity_participant_activity FOREIGN KEY (activity_id) REFERENCES Activity (ID) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_activity_participant_member FOREIGN KEY (member_id) REFERENCES `member` (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS ActivityAsset (
  id INT NOT NULL AUTO_INCREMENT,
  activity_id INT NOT NULL,
  asset_id INT NOT NULL,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY ux_activity_asset (activity_id, asset_id),
  KEY ix_activity_asset_asset (asset_id),
  CONSTRAINT fk_activity_asset_activity FOREIGN KEY (activity_id) REFERENCES Activity (ID) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_activity_asset_asset FOREIGN KEY (asset_id) REFERENCES Asset (ID) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Activity: preserve old DriverID / VehicleID links into relation tables before dropping.
SET @c_driver := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'DriverID'
);
SET @q := IF(@c_driver > 0,
  'INSERT IGNORE INTO ActivityEmployee (activity_id, employee_id) SELECT ID, DriverID FROM Activity WHERE DriverID IS NOT NULL AND DriverID > 0',
  'SELECT 1'
);
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c_vehicle := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'VehicleID'
);
SET @q := IF(@c_vehicle > 0,
  'INSERT IGNORE INTO ActivityAsset (activity_id, asset_id) SELECT ID, VehicleID FROM Activity WHERE VehicleID IS NOT NULL AND VehicleID > 0',
  'SELECT 1'
);
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Activity: drop old fields and add new fields.
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'DriverID');
SET @q := IF(@c > 0, 'ALTER TABLE Activity DROP COLUMN DriverID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'VehicleID');
SET @q := IF(@c > 0, 'ALTER TABLE Activity DROP COLUMN VehicleID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'TripDays');
SET @q := IF(@c > 0, 'ALTER TABLE Activity DROP COLUMN TripDays', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'Distance');
SET @q := IF(@c > 0, 'ALTER TABLE Activity DROP COLUMN Distance', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'StartDate');
SET @q := IF(@c > 0, 'ALTER TABLE Activity CHANGE COLUMN StartDate start_date DATETIME NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'EndDate');
SET @q := IF(@c > 0, 'ALTER TABLE Activity CHANGE COLUMN EndDate end_date DATETIME NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'start_date');
SET @q := IF(@c = 0, 'ALTER TABLE Activity ADD COLUMN start_date DATETIME NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'end_date');
SET @q := IF(@c = 0, 'ALTER TABLE Activity ADD COLUMN end_date DATETIME NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'location');
SET @q := IF(@c = 0, 'ALTER TABLE Activity ADD COLUMN location JSON NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Asset: VIN -> asset_tag, remove old fields, add description.
SET @c_vin := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Asset' AND column_name = 'VIN');
SET @c_tag := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Asset' AND column_name = 'asset_tag');
SET @q := IF(@c_vin > 0 AND @c_tag = 0, 'ALTER TABLE Asset CHANGE COLUMN VIN asset_tag VARCHAR(30) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Asset' AND column_name = 'Capacity');
SET @q := IF(@c > 0, 'ALTER TABLE Asset DROP COLUMN Capacity', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Asset' AND column_name = 'ModelInfo');
SET @q := IF(@c > 0, 'ALTER TABLE Asset DROP COLUMN ModelInfo', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Asset' AND column_name = 'GPSID');
SET @q := IF(@c > 0, 'ALTER TABLE Asset DROP COLUMN GPSID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c_desc := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Asset' AND column_name = 'description');
SET @q := IF(@c_desc = 0, 'ALTER TABLE Asset ADD COLUMN description VARCHAR(2000) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET FOREIGN_KEY_CHECKS = @OLD_FK;

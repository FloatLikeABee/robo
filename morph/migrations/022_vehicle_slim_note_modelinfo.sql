-- Vehicle: drop legacy columns; Comments -> Note; Model INT -> ModelInfo VARCHAR(100).
-- Trip.VehicleID stores Vehicle.ID (was business VehicleID); remap before dropping VehicleID/DBID.
-- MySQL 5.7+ / MariaDB.

SET @db := DATABASE();
SET FOREIGN_KEY_CHECKS = 0;

-- --- 1) Trip: VehicleID business key -> Vehicle.ID ---
SET @veh_has_business := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'VehicleID'
);

SET @s := IF(@veh_has_business > 0,
  'UPDATE Trip t INNER JOIN Vehicle v ON t.DBID = v.DBID AND t.VehicleID = v.VehicleID SET t.VehicleID = v.ID WHERE t.VehicleID IS NOT NULL AND t.VehicleID <> 0',
  'SELECT 1');
PREPARE ps FROM @s; EXECUTE ps; DEALLOCATE PREPARE ps;

-- --- 2) Drop Vehicle indexes (only if present) ---
SET @ix := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Vehicle' AND index_name = 'UQ_Vehicle_DBIDVehicleID'
);
SET @q := IF(@ix > 0, 'ALTER TABLE Vehicle DROP INDEX UQ_Vehicle_DBIDVehicleID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @ix := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Vehicle' AND index_name = 'IX_Vehicle_DBID_ComparativeAnalysis'
);
SET @q := IF(@ix > 0, 'ALTER TABLE Vehicle DROP INDEX IX_Vehicle_DBID_ComparativeAnalysis', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @ix := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Vehicle' AND index_name = 'IX_Vehicle_GPSID_VendorId'
);
SET @q := IF(@ix > 0, 'ALTER TABLE Vehicle DROP INDEX IX_Vehicle_GPSID_VendorId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- --- 3) Drop legacy columns ---
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'DBID');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN DBID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'VehicleID');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN VehicleID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'LastUpdatedID');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN LastUpdatedID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'LastUpdatedType');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN LastUpdatedType', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'FuelConsumption');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN FuelConsumption', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'EstLife');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN EstLife', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'PurchaseOdometer');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN PurchaseOdometer', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'SalvageOdometer');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN SalvageOdometer', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'SalvageValue');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN SalvageValue', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'SalvageDate');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN SalvageDate', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'MaxWeight');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN MaxWeight', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'ComparativeAnalysis');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN ComparativeAnalysis', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'VendorId');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN VendorId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'ExternalId');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN ExternalId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'CreatedOn');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN CreatedOn', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'CreatedBy');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN CreatedBy', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- --- 4) Comments -> Note ---
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'Comments');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle CHANGE COLUMN Comments Note LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- --- 5) Model -> ModelInfo VARCHAR(100) ---
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'Model');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle CHANGE COLUMN Model ModelInfo VARCHAR(100) NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET FOREIGN_KEY_CHECKS = 1;

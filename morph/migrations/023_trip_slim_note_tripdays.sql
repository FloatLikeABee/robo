-- Trip: drop legacy columns; Comments -> Note; `Day` -> TripDays.
-- MySQL 5.7+ / MariaDB.

SET @db := DATABASE();
SET FOREIGN_KEY_CHECKS = 0;

-- --- 1) Drop Trip indexes (only if present) ---
SET @ix := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Trip' AND index_name = 'AK_Trip_DBIDTripID'
);
SET @q := IF(@ix > 0, 'ALTER TABLE Trip DROP INDEX AK_Trip_DBIDTripID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @ix := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Trip' AND index_name = 'IX_Trip_DBIDAideID'
);
SET @q := IF(@ix > 0, 'ALTER TABLE Trip DROP INDEX IX_Trip_DBIDAideID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @ix := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Trip' AND index_name = 'IX_Trip_DBIDDriverID'
);
SET @q := IF(@ix > 0, 'ALTER TABLE Trip DROP INDEX IX_Trip_DBIDDriverID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @ix := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Trip' AND index_name = 'IX_Trip_DBIDVehicleID'
);
SET @q := IF(@ix > 0, 'ALTER TABLE Trip DROP INDEX IX_Trip_DBIDVehicleID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- --- 2) Drop legacy columns ---
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'DBID');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN DBID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'TripID');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN TripID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'LastUpdatedID');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN LastUpdatedID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'LastUpdatedType');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN LastUpdatedType', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'CreatedOn');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN CreatedOn', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'CreatedBy');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN CreatedBy', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'AideID');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN AideID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'FilterName');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN FilterName', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'NonDisabled');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN NonDisabled', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'Schools');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN Schools', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'Session');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN `Session`', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'HomeSchl');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN HomeSchl', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'HomeTrans');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN HomeTrans', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'Shuttle');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN Shuttle', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'NumTransport');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN NumTransport', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'MaxOnBus');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN MaxOnBus', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'StartTime');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN StartTime', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'FinishTime');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN FinishTime', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'TripAlias');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN TripAlias', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'Cost');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN Cost', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'iShow');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN iShow', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'iName');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN iName', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'DHDistance');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN DHDistance', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'FilterSpec');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN FilterSpec', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'GPSEnabledFlag');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN GPSEnabledFlag', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'TravelScenarioId');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN TravelScenarioId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'IntGratChar1');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN IntGratChar1', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'IntGratChar2');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN IntGratChar2', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'IntGratDate1');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN IntGratDate1', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'IntGratDate2');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN IntGratDate2', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'IntGratNum1');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN IntGratNum1', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'IntGratNum2');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN IntGratNum2', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'StartDate');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN StartDate', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'EndDate');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN EndDate', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'SpeedType');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN SpeedType', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'DefaultSpeed');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN DefaultSpeed', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'ExcludeNoStudStopAndDirections');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN ExcludeNoStudStopAndDirections', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'RouteId');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN RouteId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'DHDuration');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN DHDuration', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'DistanceWithPassengers');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN DistanceWithPassengers', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'DurationWithPassengers');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN DurationWithPassengers', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'AverageRideTime');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN AverageRideTime', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'MaxRideTime');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN MaxRideTime', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- --- 3) Comments -> Note ---
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'Comments');
SET @q := IF(@c > 0, 'ALTER TABLE Trip CHANGE COLUMN Comments Note LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- --- 4) Day -> TripDays ---
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'Day');
SET @q := IF(@c > 0, 'ALTER TABLE Trip CHANGE COLUMN `Day` TripDays TINYINT NULL DEFAULT 0', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET FOREIGN_KEY_CHECKS = 1;

-- Staff: drop legacy columns; Comments -> Note; address via CityId, PostalCodeId, StateId (like Student).
-- StaffStaffType: StaffID becomes FK to Staff.ID (was business key with DBID).
-- Trip: DriverID / AideID remapped from legacy Staff.StaffID to Staff.ID before StaffID column is dropped.
-- MySQL 5.7+ / MariaDB (no DROP COLUMN IF EXISTS).

SET @db := DATABASE();
SET FOREIGN_KEY_CHECKS = 0;

-- --- 1) StaffStaffType: migrate off (DBID, business StaffID) to Staff.ID ---
SET @sst_has_dbid := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'StaffStaffType' AND column_name = 'DBID'
);

SET @s := IF(@sst_has_dbid > 0,
  'DROP TABLE IF EXISTS StaffStaffType_new',
  'SELECT 1');
PREPARE ps FROM @s; EXECUTE ps; DEALLOCATE PREPARE ps;

SET @s := IF(@sst_has_dbid > 0,
  'CREATE TABLE StaffStaffType_new (
    ID INT NOT NULL AUTO_INCREMENT,
    StaffID INT NOT NULL,
    StaffTypeID INT NOT NULL,
    PrimaryFlag TINYINT(1) NULL DEFAULT 0,
    PRIMARY KEY (ID),
    UNIQUE KEY UQ_StaffStaffType_Staff_StaffType (StaffID, StaffTypeID),
    KEY IX_StaffStaffType_StaffTypeID (StaffTypeID),
    CONSTRAINT FK_StaffStaffType_Staff FOREIGN KEY (StaffID) REFERENCES Staff (ID) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_StaffStaffType_StaffType FOREIGN KEY (StaffTypeID) REFERENCES StaffType (StaffTypeID)
  )',
  'SELECT 1');
PREPARE ps FROM @s; EXECUTE ps; DEALLOCATE PREPARE ps;

SET @s := IF(@sst_has_dbid > 0,
  'INSERT INTO StaffStaffType_new (StaffID, StaffTypeID, PrimaryFlag)
   SELECT s.ID, sst.StaffTypeID, sst.PrimaryFlag
   FROM StaffStaffType sst
   INNER JOIN Staff s ON sst.DBID = s.DBID AND sst.StaffID = s.StaffID',
  'SELECT 1');
PREPARE ps FROM @s; EXECUTE ps; DEALLOCATE PREPARE ps;

SET @s := IF(@sst_has_dbid > 0, 'DROP TABLE StaffStaffType', 'SELECT 1');
PREPARE ps FROM @s; EXECUTE ps; DEALLOCATE PREPARE ps;

SET @s := IF(@sst_has_dbid > 0, 'RENAME TABLE StaffStaffType_new TO StaffStaffType', 'SELECT 1');
PREPARE ps FROM @s; EXECUTE ps; DEALLOCATE PREPARE ps;

-- --- 2) Trip: remap driver / aide from business StaffID to Staff.ID ---
SET @staff_has_business_id := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'StaffID'
);

SET @s := IF(@staff_has_business_id > 0,
  'UPDATE Trip t INNER JOIN Staff s ON t.DBID = s.DBID AND t.DriverID = s.StaffID SET t.DriverID = s.ID WHERE t.DriverID IS NOT NULL AND t.DriverID <> 0',
  'SELECT 1');
PREPARE ps FROM @s; EXECUTE ps; DEALLOCATE PREPARE ps;

SET @s := IF(@staff_has_business_id > 0,
  'UPDATE Trip t INNER JOIN Staff s ON t.DBID = s.DBID AND t.AideID = s.StaffID SET t.AideID = s.ID WHERE t.AideID IS NOT NULL AND t.AideID <> 0',
  'SELECT 1');
PREPARE ps FROM @s; EXECUTE ps; DEALLOCATE PREPARE ps;

-- --- 3) Drop Staff indexes (only if present) ---
SET @ix := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Staff' AND index_name = 'UQ_Staff_DBIDStaffID'
);
SET @q := IF(@ix > 0, 'ALTER TABLE Staff DROP INDEX UQ_Staff_DBIDStaffID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @ix := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = @db AND table_name = 'Staff' AND index_name = 'IX_Staff_LicenseNumber'
);
SET @q := IF(@ix > 0, 'ALTER TABLE Staff DROP INDEX IX_Staff_LicenseNumber', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- --- 4) Drop legacy Staff columns (one at a time) ---
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'DBID');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN DBID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'StaffID');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN StaffID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'CreatedOn');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN CreatedOn', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'CreatedBy');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN CreatedBy', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'StaffLocalID');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN StaffLocalID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'MailStreet1');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN MailStreet1', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'MailStreet2');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN MailStreet2', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'MailCity');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN MailCity', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'MailZip');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN MailZip', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'HomePhone');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN HomePhone', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'WorkPhone');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN WorkPhone', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'LicenseNumber');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN LicenseNumber', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'LicenseState');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN LicenseState', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'LicenseExpiration');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN LicenseExpiration', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'LicenseClass');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN LicenseClass', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'LicenseEndorsements');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN LicenseEndorsements', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'LicenseRestrictions');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN LicenseRestrictions', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'HireDate');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN HireDate', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'Rate');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN Rate', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'OTRate');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN OTRate', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'LastUpdated');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN LastUpdated', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'LastUpdatedID');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN LastUpdatedID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'LastUpdatedType');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN LastUpdatedType', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'DeletedFlag');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN DeletedFlag', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'DeletedDate');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN DeletedDate', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'ApplicationField');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN ApplicationField', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'FingerPrint');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN FingerPrint', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'SuperintendentApprov');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN SuperintendentApprov', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'NewHireOrient');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN NewHireOrient', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'Abstract');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN Abstract', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'Interview');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN Interview', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'DefensiveDriving');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN DefensiveDriving', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'DrivingTestPractical');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN DrivingTestPractical', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'DrivingTestWritten');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN DrivingTestWritten', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'MedicalExam');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN MedicalExam', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'PPTField');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN PPTField', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'HepatitisB');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN HepatitisB', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'Certification');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN Certification', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'BasicField');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN BasicField', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'Advanced');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN Advanced', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'PreService');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN PreService', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'HandicapPreService');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN HandicapPreService', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'RefresherPart1');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN RefresherPart1', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'RefresherPart2');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN RefresherPart2', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'HandicapRef');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN HandicapRef', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'DistrictID');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN DistrictID', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'Mail_State_Id');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN Mail_State_Id', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- --- 5) Comments -> Note ---
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'Comments');
SET @q := IF(@c > 0, 'ALTER TABLE Staff CHANGE COLUMN Comments Note LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- --- 6) Address reference columns + FKs ---
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'CityId');
SET @q := IF(@c = 0, 'ALTER TABLE Staff ADD COLUMN CityId INT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'PostalCodeId');
SET @q := IF(@c = 0, 'ALTER TABLE Staff ADD COLUMN PostalCodeId INT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'StateId');
SET @q := IF(@c = 0, 'ALTER TABLE Staff ADD COLUMN StateId INT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @fk := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Staff' AND constraint_name = 'FK_Staff_MailingCity'
);
SET @q := IF(@fk = 0,
  'ALTER TABLE Staff ADD CONSTRAINT FK_Staff_MailingCity FOREIGN KEY (CityId) REFERENCES MailingCity (ID) ON DELETE SET NULL ON UPDATE CASCADE',
  'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @fk := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Staff' AND constraint_name = 'FK_Staff_MailingPostalCode'
);
SET @q := IF(@fk = 0,
  'ALTER TABLE Staff ADD CONSTRAINT FK_Staff_MailingPostalCode FOREIGN KEY (PostalCodeId) REFERENCES MailingPostalCode (ID) ON DELETE SET NULL ON UPDATE CASCADE',
  'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @fk := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Staff' AND constraint_name = 'FK_Staff_MailingState'
);
SET @q := IF(@fk = 0,
  'ALTER TABLE Staff ADD CONSTRAINT FK_Staff_MailingState FOREIGN KEY (StateId) REFERENCES MailingState (ID) ON DELETE SET NULL ON UPDATE CASCADE',
  'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET FOREIGN_KEY_CHECKS = 1;

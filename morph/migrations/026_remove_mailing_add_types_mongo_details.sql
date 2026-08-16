-- Remove city/postal/state + mailing ref tables, mailing fields on contact/district; add *Type columns;
-- remove note/detail columns (content moves to MongoDB collection entity_details in application).
SET @db := DATABASE();
SET FOREIGN_KEY_CHECKS = 0;

-- School: drop FKs to mailing tables, then address columns
SET @fk := (SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'School' AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'FK_School_MailingCity');
SET @q := IF(@fk > 0, 'ALTER TABLE School DROP FOREIGN KEY FK_School_MailingCity', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @fk := (SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'School' AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'FK_School_MailingPostalCode');
SET @q := IF(@fk > 0, 'ALTER TABLE School DROP FOREIGN KEY FK_School_MailingPostalCode', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @fk := (SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'School' AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'FK_School_MailingState');
SET @q := IF(@fk > 0, 'ALTER TABLE School DROP FOREIGN KEY FK_School_MailingState', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'School' AND column_name = 'CityId');
SET @q := IF(@c > 0, 'ALTER TABLE School DROP COLUMN CityId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'School' AND column_name = 'PostalCodeId');
SET @q := IF(@c > 0, 'ALTER TABLE School DROP COLUMN PostalCodeId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'School' AND column_name = 'StateId');
SET @q := IF(@c > 0, 'ALTER TABLE School DROP COLUMN StateId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Student: drop address FKs
SET @fk := (SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Student' AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'FK_Student_MailingCity');
SET @q := IF(@fk > 0, 'ALTER TABLE Student DROP FOREIGN KEY FK_Student_MailingCity', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @fk := (SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Student' AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'FK_Student_MailingPostalCode');
SET @q := IF(@fk > 0, 'ALTER TABLE Student DROP FOREIGN KEY FK_Student_MailingPostalCode', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @fk := (SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Student' AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'FK_Student_MailingState');
SET @q := IF(@fk > 0, 'ALTER TABLE Student DROP FOREIGN KEY FK_Student_MailingState', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'CityId');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN CityId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'PostalCodeId');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN PostalCodeId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'StateId');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN StateId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Staff: drop address FKs
SET @fk := (SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Staff' AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'FK_Staff_MailingCity');
SET @q := IF(@fk > 0, 'ALTER TABLE Staff DROP FOREIGN KEY FK_Staff_MailingCity', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @fk := (SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Staff' AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'FK_Staff_MailingPostalCode');
SET @q := IF(@fk > 0, 'ALTER TABLE Staff DROP FOREIGN KEY FK_Staff_MailingPostalCode', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @fk := (SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE table_schema = @db AND table_name = 'Staff' AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'FK_Staff_MailingState');
SET @q := IF(@fk > 0, 'ALTER TABLE Staff DROP FOREIGN KEY FK_Staff_MailingState', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'CityId');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN CityId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'PostalCodeId');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN PostalCodeId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'StateId');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN StateId', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- District: mailing
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'District' AND column_name = 'MailCity');
SET @q := IF(@c > 0, 'ALTER TABLE District DROP COLUMN MailCity', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'District' AND column_name = 'MailStreet1');
SET @q := IF(@c > 0, 'ALTER TABLE District DROP COLUMN MailStreet1', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'District' AND column_name = 'MailStreet2');
SET @q := IF(@c > 0, 'ALTER TABLE District DROP COLUMN MailStreet2', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'District' AND column_name = 'MailZip');
SET @q := IF(@c > 0, 'ALTER TABLE District DROP COLUMN MailZip', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- contact: mailing
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

-- Add type columns
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'EmployType');
SET @q := IF(@c = 0, "ALTER TABLE Staff ADD COLUMN EmployType VARCHAR(64) NULL AFTER ActiveFlag", 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'AssetType');
SET @q := IF(@c = 0, "ALTER TABLE Vehicle ADD COLUMN AssetType VARCHAR(64) NULL AFTER AssetID", 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'ActivityType');
SET @q := IF(@c = 0, "ALTER TABLE Trip ADD COLUMN ActivityType VARCHAR(64) NULL", 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'School' AND column_name = 'FacilityType');
SET @q := IF(@c = 0, "ALTER TABLE School ADD COLUMN FacilityType VARCHAR(64) NULL AFTER Private", 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'ParticipantType');
SET @q := IF(@c = 0, "ALTER TABLE Student ADD COLUMN ParticipantType VARCHAR(64) NULL", 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Remove note columns (details stored in MongoDB)
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Student' AND column_name = 'Note');
SET @q := IF(@c > 0, 'ALTER TABLE Student DROP COLUMN Note', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'School' AND column_name = 'Notes');
SET @q := IF(@c > 0, 'ALTER TABLE School DROP COLUMN Notes', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Staff' AND column_name = 'Note');
SET @q := IF(@c > 0, 'ALTER TABLE Staff DROP COLUMN Note', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'Note');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN Note', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Vehicle' AND column_name = 'VehicleDetail');
SET @q := IF(@c > 0, 'ALTER TABLE Vehicle DROP COLUMN VehicleDetail', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Trip' AND column_name = 'Note');
SET @q := IF(@c > 0, 'ALTER TABLE Trip DROP COLUMN Note', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'Note');
SET @q := IF(@c > 0, 'ALTER TABLE `contact` DROP COLUMN Note', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'District' AND column_name = 'Comments');
SET @q := IF(@c > 0, 'ALTER TABLE District DROP COLUMN Comments', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Drop reference tables
DROP TABLE IF EXISTS MailingState;
DROP TABLE IF EXISTS MailingPostalCode;
DROP TABLE IF EXISTS MailingCity;

SET FOREIGN_KEY_CHECKS = 1;

-- Upgrade legacy databases only. Skip if you initialized from the current migrations/001_schema_mysql.sql
-- (already has varchar Mail_City / Mail_Street / Mail_Zip and no FormResult table).
--
-- School: Geo_* + int Mail_* -> varchar Mail_City, Mail_Street, Mail_Zip; drop Mail_Street1
-- District: drop Mail_State_Id, LastUpdated, LastUpdatedBy
-- record_contact: drop SortOrder, LastUpdated
-- Vehicle: drop Bus_Num, InActive, LastUpdated
-- Drop FormResult table

ALTER TABLE School
  DROP COLUMN Mail_City,
  DROP COLUMN Mail_Zip,
  DROP COLUMN Mail_Street1,
  CHANGE COLUMN Geo_City Mail_City VARCHAR(30) NULL,
  CHANGE COLUMN Geo_Street Mail_Street VARCHAR(150) NULL,
  CHANGE COLUMN Geo_Zip Mail_Zip VARCHAR(20) NULL;

ALTER TABLE District
  DROP COLUMN Mail_State_Id,
  DROP COLUMN LastUpdated,
  DROP COLUMN LastUpdatedBy;

ALTER TABLE record_contact
  DROP COLUMN SortOrder,
  DROP COLUMN LastUpdated;

ALTER TABLE Vehicle
  DROP COLUMN Bus_Num,
  DROP COLUMN InActive,
  DROP COLUMN LastUpdated;

DROP TABLE IF EXISTS FormResult;

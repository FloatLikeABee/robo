-- School: remove legacy columns; Comments -> Notes; mailing address via MailingCity / MailingPostalCode / MailingState (CityId, PostalCodeId, StateId).

SET FOREIGN_KEY_CHECKS = 0;

ALTER TABLE School DROP INDEX UQ_School_DBIDSchoolCode;

ALTER TABLE School
  DROP COLUMN DBID,
  DROP COLUMN SchoolID,
  DROP COLUMN Mail_City,
  DROP COLUMN Mail_Street,
  DROP COLUMN Mail_Zip,
  DROP COLUMN Geo_County,
  DROP COLUMN Mail_Street2,
  DROP COLUMN Phone,
  DROP COLUMN Begin_time,
  DROP COLUMN End_time,
  DROP COLUMN ArrivalTime,
  DROP COLUMN DepartTime,
  DROP COLUMN TSchl,
  DROP COLUMN GeoConfidence,
  DROP COLUMN DispGrade,
  DROP COLUMN LastUpdated,
  DROP COLUMN LastUpdatedID,
  DROP COLUMN LastUpdatedType,
  DROP COLUMN Mail_State_Id,
  DROP COLUMN CreatedOn,
  DROP COLUMN CreatedBy;

ALTER TABLE School CHANGE COLUMN Comments Notes LONGTEXT NULL;

ALTER TABLE School
  ADD COLUMN CityId INT NULL,
  ADD COLUMN PostalCodeId INT NULL,
  ADD COLUMN StateId INT NULL;

ALTER TABLE School ADD UNIQUE KEY UQ_School_SchoolCode (SchoolCode);

ALTER TABLE School
  ADD CONSTRAINT FK_School_MailingCity FOREIGN KEY (CityId) REFERENCES MailingCity (ID) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT FK_School_MailingPostalCode FOREIGN KEY (PostalCodeId) REFERENCES MailingPostalCode (ID) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT FK_School_MailingState FOREIGN KEY (StateId) REFERENCES MailingState (ID) ON DELETE SET NULL ON UPDATE CASCADE;

SET FOREIGN_KEY_CHECKS = 1;

-- Contact: add mailing address and keep Comments as the note column.
-- Contact is used for school personnel, student parents, and any entity's contacts.
ALTER TABLE `contact`
  ADD COLUMN Mail_Street1 VARCHAR(150) NULL AFTER Mobile,
  ADD COLUMN Mail_Street2 VARCHAR(150) NULL AFTER Mail_Street1,
  ADD COLUMN Mail_City INT NULL AFTER Mail_Street2,
  ADD COLUMN Mail_Zip INT NULL AFTER Mail_City,
  ADD COLUMN Mail_State_Id INT NULL AFTER Mail_Zip;

-- RecordContact: links any entity (student, school, trip, vehicle, staff, district) to one or more contacts.
-- EntityType: 'student' | 'school' | 'trip' | 'vehicle' | 'staff' | 'district'
-- Relationship: user-defined label, e.g. "Parent", "Guardian", "Principal", "Emergency contact"
CREATE TABLE IF NOT EXISTS `record_contact` (
  ID INT NOT NULL AUTO_INCREMENT,
  DBID INT NOT NULL,
  EntityType VARCHAR(20) NOT NULL,
  RecordID INT NOT NULL,
  ContactID INT NOT NULL,
  Relationship VARCHAR(100) NULL,
  IsPrimary TINYINT(1) NOT NULL DEFAULT 0,
  CreatedOn DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  KEY IX_record_contact_entity (DBID, EntityType, RecordID),
  KEY IX_record_contact_contact (ContactID)
);

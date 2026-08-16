-- Comments for main data records (students, staff, schools, vehicles, trips).
-- One record can have many comments; one comment can have many sub-comments (ParentID).
CREATE TABLE IF NOT EXISTS `comment` (
  ID INT NOT NULL AUTO_INCREMENT,
  EntityType VARCHAR(20) NOT NULL,
  RecordID INT NOT NULL,
  ParentID INT NULL,
  AuthorUserID INT NULL,
  Body LONGTEXT NOT NULL,
  CreatedOn DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  LastUpdated DATETIME NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  KEY IX_comment_entity (EntityType, RecordID),
  KEY IX_comment_parent (ParentID)
);

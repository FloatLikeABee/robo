-- Tools module: notes/messages, email broadcast history.
-- User table reference: User.UserID (existing table).

-- tool_note: private notes and public messages (to self, to a user, or public).
-- AuthorUserID = who wrote it. TargetType: 'self' | 'user' | 'public'.
-- TargetUserID = recipient when TargetType = 'user'. ReadAt = when recipient read it.
CREATE TABLE IF NOT EXISTS `tool_note` (
  ID INT NOT NULL AUTO_INCREMENT,
  AuthorUserID INT NOT NULL,
  TargetType VARCHAR(10) NOT NULL DEFAULT 'self',
  TargetUserID INT NULL,
  IsPrivate TINYINT(1) NOT NULL DEFAULT 1,
  Title VARCHAR(255) NULL,
  Body LONGTEXT NULL,
  ReadAt DATETIME NULL,
  CreatedOn DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  LastUpdated DATETIME NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  KEY IX_tool_note_author (AuthorUserID),
  KEY IX_tool_note_target (TargetType, TargetUserID),
  KEY IX_tool_note_created (CreatedOn)
);

-- tool_broadcast: email broadcast history (sent campaigns).
CREATE TABLE IF NOT EXISTS `tool_broadcast` (
  ID INT NOT NULL AUTO_INCREMENT,
  SentByUserID INT NOT NULL,
  Subject VARCHAR(500) NOT NULL,
  Body LONGTEXT NULL,
  RecipientCount INT NOT NULL DEFAULT 0,
  CreatedOn DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  KEY IX_tool_broadcast_sent_by (SentByUserID)
);

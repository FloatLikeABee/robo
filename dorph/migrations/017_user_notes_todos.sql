-- Personal notes and TODOs per user (User.UserID). Default user 1 until auth is wired.
CREATE TABLE IF NOT EXISTS `user_note_todo` (
  ID INT NOT NULL AUTO_INCREMENT,
  UserID INT NOT NULL DEFAULT 1,
  ItemType ENUM('note', 'todo') NOT NULL,
  Title VARCHAR(500) NULL,
  Body LONGTEXT NULL,
  Completed TINYINT(1) NOT NULL DEFAULT 0,
  DeadlineAt DATETIME NULL,
  CreatedOn DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  LastUpdated DATETIME NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  KEY IX_user_note_todo_user (UserID, ItemType, CreatedOn),
  KEY IX_user_note_todo_deadline (UserID, ItemType, DeadlineAt)
);

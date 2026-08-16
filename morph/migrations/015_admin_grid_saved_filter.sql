-- Shared saved filter definitions for admin grids (all users)
CREATE TABLE IF NOT EXISTS AdminGridSavedFilter (
  ID INT NOT NULL AUTO_INCREMENT,
  GridKey VARCHAR(80) NOT NULL,
  Name VARCHAR(120) NOT NULL,
  FilterJSON LONGTEXT NOT NULL,
  CreatedOn DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  UNIQUE KEY UIX_AdminGridSavedFilter_GridKey_Name (GridKey, Name),
  KEY IX_AdminGridSavedFilter_GridKey (GridKey)
);

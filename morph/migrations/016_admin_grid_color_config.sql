-- Shared row color tag config per admin grid (presentation only; all users)
CREATE TABLE IF NOT EXISTS AdminGridColorConfig (
  GridKey VARCHAR(80) NOT NULL,
  ConfigJSON LONGTEXT NOT NULL,
  UpdatedOn DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (GridKey)
);

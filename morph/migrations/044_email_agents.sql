-- AI Email Agent: monitor mailboxes, screen noise, auto-reply to useful mail.
CREATE TABLE IF NOT EXISTS `email_agent` (
  ID INT NOT NULL AUTO_INCREMENT,
  OwnerUserID INT NOT NULL,
  Name VARCHAR(255) NOT NULL,
  Enabled TINYINT(1) NOT NULL DEFAULT 0,
  ImapHost VARCHAR(255) NULL,
  ImapPort INT NOT NULL DEFAULT 993,
  ImapUser VARCHAR(255) NULL,
  ImapPassword VARCHAR(512) NULL,
  SmtpHost VARCHAR(255) NULL,
  SmtpPort INT NOT NULL DEFAULT 587,
  SmtpUser VARCHAR(255) NULL,
  SmtpPassword VARCHAR(512) NULL,
  MailboxFolder VARCHAR(128) NOT NULL DEFAULT 'INBOX',
  PollIntervalSec INT NOT NULL DEFAULT 300,
  ScreeningPrompt LONGTEXT NULL,
  ReplyPrompt LONGTEXT NULL,
  LastCheckedOn DATETIME NULL,
  CreatedOn DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  LastUpdated DATETIME NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  KEY IX_email_agent_owner (OwnerUserID),
  KEY IX_email_agent_enabled (Enabled)
);

CREATE TABLE IF NOT EXISTS `email_agent_log` (
  ID INT NOT NULL AUTO_INCREMENT,
  AgentID INT NOT NULL,
  MessageUID VARCHAR(128) NULL,
  Subject VARCHAR(512) NULL,
  FromAddress VARCHAR(512) NULL,
  Action VARCHAR(32) NOT NULL,
  Detail LONGTEXT NULL,
  CreatedOn DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  KEY IX_email_agent_log_agent (AgentID),
  KEY IX_email_agent_log_created (CreatedOn)
);

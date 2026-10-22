-- Broadcast attachment metadata (ComposerX files, FormsX links, DataX publishes, custom URLs).
ALTER TABLE `tool_broadcast`
  ADD COLUMN `AttachmentsJSON` LONGTEXT NULL AFTER `RecipientCount`;

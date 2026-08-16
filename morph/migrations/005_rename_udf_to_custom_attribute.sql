-- One-time migration: rename legacy UDF / UDFType tables to CustomAttribute / CustomAttributeType.
-- Run on databases that were created from an older 001_schema_mysql.sql (before the CustomAttribute rename).
-- If your database was created from the updated schema, skip this file (tables are already named CustomAttribute).

-- MySQL 5.7+

RENAME TABLE `UDF` TO `CustomAttribute`;
RENAME TABLE `UDFType` TO `CustomAttributeType`;

ALTER TABLE `CustomAttribute` DROP INDEX `UIX_UDF_DataTypeDisplayName`;
ALTER TABLE `CustomAttribute` ADD UNIQUE KEY `UIX_CustomAttribute_DataTypeDisplayName` (`DataType`, `DisplayName`);

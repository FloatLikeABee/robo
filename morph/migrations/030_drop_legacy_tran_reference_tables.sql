-- Drop legacy lookup / merge / schedule tables (no longer used; custom attributes use numeric DataType column, not this table).
-- ClickUp: remove Grade, MergeDocument, MergeDocumentSent, SchoolGrade, StudentSchedule, Form, DataType, Disability_Codes, Ethnic_Codes
SET @OLD_FK := @@FOREIGN_KEY_CHECKS;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `StudentSchedule`;
DROP TABLE IF EXISTS `SchoolGrade`;
DROP TABLE IF EXISTS `Grade`;
DROP TABLE IF EXISTS `MergeDocumentSent`;
DROP TABLE IF EXISTS `MergeDocument`;
DROP TABLE IF EXISTS `Form`;
DROP TABLE IF EXISTS `Ethnic_Codes`;
DROP TABLE IF EXISTS `Disability_Codes`;
DROP TABLE IF EXISTS `DataType`;

SET FOREIGN_KEY_CHECKS = @OLD_FK;

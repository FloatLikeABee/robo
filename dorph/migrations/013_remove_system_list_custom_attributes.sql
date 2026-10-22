-- Remove previously seeded system list custom attributes (Disability Codes, Ethnic Codes, School Grades).

USE tran;

DELETE FROM CustomAttribute
WHERE DisplayName IN ('Disability Codes', 'Ethnic Codes', 'School Grades');

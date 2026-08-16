-- Values for custom attributes per entity record (efficient lookup by record + data type scope via CustomAttribute)
CREATE TABLE IF NOT EXISTS CustomAttributeValue (
  ID INT NOT NULL AUTO_INCREMENT,
  CustomAttributeID INT NOT NULL,
  RecordID INT NOT NULL,
  ValueText LONGTEXT NULL,
  PRIMARY KEY (ID),
  UNIQUE KEY UIX_CustomAttributeValue_AttrRecord (CustomAttributeID, RecordID),
  KEY IX_CustomAttributeValue_Record (RecordID),
  KEY IX_CustomAttributeValue_Attr (CustomAttributeID)
);

-- Contact entity type for custom attributes (admin Contacts grid)
INSERT IGNORE INTO DataType (ID, Type, SystemDefined, DisplayInNav) VALUES
  (7, 'Contact', 1, 1);

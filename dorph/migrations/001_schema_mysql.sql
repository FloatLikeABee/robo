-- TranDemo: Schools/District Trip Info Management - MySQL Schema
-- Adapted from SQL Server DDL. UDGrid -> Form.

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE DATABASE IF NOT EXISTS tran DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE tran;

-- DataType
CREATE TABLE DataType (
  ID SMALLINT NOT NULL,
  `Type` VARCHAR(100) NULL,
  SystemDefined TINYINT(1) NOT NULL DEFAULT 0,
  DisplayInNav TINYINT(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (ID)
);

-- Disability_Codes
CREATE TABLE Disability_Codes (
  ID INT NOT NULL AUTO_INCREMENT,
  DBID INT NOT NULL,
  DisCodeID INT NOT NULL,
  Code VARCHAR(4) NULL,
  Comments VARCHAR(128) NULL,
  Description VARCHAR(32) NULL,
  LoadTimePerStudent INT NULL DEFAULT 0,
  PRIMARY KEY (ID),
  UNIQUE KEY UQ_DisabilityCodes_DBIDDisCodeID (DBID, DisCodeID)
);

-- District
CREATE TABLE District (
  ID INT NOT NULL AUTO_INCREMENT,
  DBID INT NOT NULL,
  DistrictID INT NOT NULL,
  District VARCHAR(4) NOT NULL,
  Name VARCHAR(30) NULL,
  Comments LONGTEXT NULL,
  MailCity INT NULL,
  MailStreet1 VARCHAR(150) NULL,
  MailStreet2 VARCHAR(150) NULL,
  MailZip INT NULL,
  LastUpdatedType SMALLINT NULL DEFAULT 0,
  CreatedOn DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  CreatedBy INT NULL,
  PRIMARY KEY (ID),
  UNIQUE KEY UQ_District_DBIDDistrict (DBID, District),
  UNIQUE KEY UQ_District_DBIDDistrictID (DBID, DistrictID)
);

-- Ethnic_Codes
CREATE TABLE Ethnic_Codes (
  ID INT NOT NULL AUTO_INCREMENT,
  DBID INT NOT NULL,
  EthnicCodeID INT NOT NULL,
  Code VARCHAR(4) NOT NULL,
  Description VARCHAR(32) NULL,
  Comments VARCHAR(128) NULL,
  PRIMARY KEY (ID),
  UNIQUE KEY UQ_EthnicCodes_DBIDEthnicCodeID (DBID, EthnicCodeID)
);

-- Grade
CREATE TABLE Grade (
  ID SMALLINT NOT NULL AUTO_INCREMENT,
  Code VARCHAR(2) NOT NULL,
  Name VARCHAR(30) NULL,
  FeedTo SMALLINT NULL,
  IsGraduatedGrade TINYINT(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (ID),
  UNIQUE KEY UIX_Grade_Code (Code)
);

-- MailingCity
CREATE TABLE MailingCity (
  ID INT NOT NULL AUTO_INCREMENT,
  Name VARCHAR(255) NULL,
  PRIMARY KEY (ID),
  UNIQUE KEY UQ_MailingCity_Name (Name)
);

-- MailingPostalCode
CREATE TABLE MailingPostalCode (
  ID INT NOT NULL AUTO_INCREMENT,
  Postal VARCHAR(20) NULL,
  PRIMARY KEY (ID),
  UNIQUE KEY UQ_MailingPostalCode_Postal (Postal)
);

-- MailingState
CREATE TABLE MailingState (
  ID INT NOT NULL AUTO_INCREMENT,
  Name VARCHAR(255) NULL,
  PRIMARY KEY (ID),
  UNIQUE KEY UQ_MailingState_Name (Name)
);

-- MergeDocument
CREATE TABLE MergeDocument (
  ID INT NOT NULL AUTO_INCREMENT,
  DataType SMALLINT NOT NULL,
  TemplateType SMALLINT NOT NULL,
  Name VARCHAR(30) NOT NULL,
  Description VARCHAR(200) NULL,
  Content LONGTEXT NULL,
  Subject LONGTEXT NULL,
  CreatedBy INT NOT NULL,
  LastUpdated DATETIME NOT NULL,
  LastUpdatedType SMALLINT NULL DEFAULT 0,
  HasHeader TINYINT(1) NULL,
  HasFooter TINYINT(1) NULL,
  HeaderContent LONGTEXT NULL,
  FooterContent LONGTEXT NULL,
  MarginTop REAL NULL,
  MarginBottom REAL NULL,
  RowSpacing REAL NULL,
  `Rows` SMALLINT NULL,
  HeaderHeight REAL NULL,
  FooterHeight REAL NULL,
  PageWidth REAL NULL,
  PageHeight REAL NULL,
  MarginRight REAL NULL,
  MarginLeft REAL NULL,
  CellWidth REAL NULL,
  CellHeight REAL NULL,
  ColumnSpacing REAL NULL,
  CellPadding REAL NULL,
  Columns SMALLINT NULL,
  MergeDocumentLibraryID INT NULL,
  CreatedOn DATETIME NULL,
  PRIMARY KEY (ID)
);

-- MergeDocumentSent
CREATE TABLE MergeDocumentSent (
  Id INT NOT NULL AUTO_INCREMENT,
  Name VARCHAR(30) NOT NULL,
  Description VARCHAR(200) NULL,
  SentBy INT NOT NULL,
  SentOn DATETIME NOT NULL,
  DataType SMALLINT NULL,
  TemplateType SMALLINT NULL,
  RunType VARCHAR(30) NULL,
  PRIMARY KEY (Id),
  KEY IX_MergeDocumentSent_SentOn (SentOn)
);

-- ScheduledJob
CREATE TABLE ScheduledJob (
  ID INT NOT NULL AUTO_INCREMENT,
  ScheduleID INT NULL,
  ScheduleDefinitionID INT NULL,
  ScheduleType VARCHAR(30) NULL,
  IsOnDemand TINYINT(1) NOT NULL,
  IsCompleted TINYINT(1) NOT NULL,
  IsExecuting TINYINT(1) NOT NULL,
  StateText VARCHAR(30) NULL,
  PlannedExecuteTime DATETIME NULL,
  CreatedOn DATETIME NOT NULL,
  LastStateUpdatedOn DATETIME NULL,
  RetriedCount INT NOT NULL,
  ExecutionOutput LONGTEXT NULL,
  LastUpdatedBy INT NULL,
  CorrelationId VARCHAR(40) NULL,
  OutputResult LONGTEXT NULL,
  PRIMARY KEY (ID),
  KEY IX_ScheduledJob_Type_OnDemand (ScheduleType, IsOnDemand),
  KEY IX_ScheduledJob_Type_Schedule_Completed_Time (ScheduleType, ScheduleID, IsOnDemand, IsCompleted, PlannedExecuteTime)
);

-- School
CREATE TABLE School (
  ID INT NOT NULL AUTO_INCREMENT,
  SchoolCode VARCHAR(10) NOT NULL,
  Name VARCHAR(150) NULL,
  Notes LONGTEXT NULL,
  DistrictID INT NULL,
  Private TINYINT(1) NOT NULL DEFAULT 0,
  XCoord DECIMAL(9,6) NULL DEFAULT 0,
  YCoord DECIMAL(9,6) NULL DEFAULT 0,
  FeedSchoolCode VARCHAR(10) NULL,
  GUID VARCHAR(50) NULL,
  Capacity INT NULL DEFAULT 0,
  CityId INT NULL,
  PostalCodeId INT NULL,
  StateId INT NULL,
  PRIMARY KEY (ID),
  UNIQUE KEY UQ_School_SchoolCode (SchoolCode),
  KEY IX_School_XCoordYCoord (XCoord, YCoord),
  CONSTRAINT FK_School_MailingCity FOREIGN KEY (CityId) REFERENCES MailingCity (ID) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT FK_School_MailingPostalCode FOREIGN KEY (PostalCodeId) REFERENCES MailingPostalCode (ID) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT FK_School_MailingState FOREIGN KEY (StateId) REFERENCES MailingState (ID) ON DELETE SET NULL ON UPDATE CASCADE
);

-- SchoolGrade
CREATE TABLE SchoolGrade (
  ID INT NOT NULL AUTO_INCREMENT,
  DBID INT NOT NULL,
  SchoolCode VARCHAR(10) NOT NULL,
  GradeID SMALLINT NOT NULL,
  PRIMARY KEY (ID)
);

-- Staff (slim: identity, note, address refs; Trip.DriverID / AideID reference Staff.ID)
CREATE TABLE Staff (
  ID INT NOT NULL AUTO_INCREMENT,
  LastName VARCHAR(50) NOT NULL,
  FirstName VARCHAR(50) NULL,
  MiddleName VARCHAR(50) NULL,
  StaffGUID VARCHAR(50) NULL,
  ActiveFlag TINYINT(1) NOT NULL DEFAULT 1,
  InactiveDate DATE NULL,
  ContractorID INT NULL DEFAULT 0,
  EMail VARCHAR(100) NULL,
  MailCounty VARCHAR(25) NULL,
  CellPhone VARCHAR(30) NULL,
  DateOfBirth DATE NULL,
  Gender SMALLINT NULL,
  Note LONGTEXT NULL,
  EmployeeID VARCHAR(50) NULL,
  UserID INT NULL,
  CityId INT NULL,
  PostalCodeId INT NULL,
  StateId INT NULL,
  PRIMARY KEY (ID),
  KEY IX_Staff_ContractorID (ContractorID),
  CONSTRAINT FK_Staff_MailingCity FOREIGN KEY (CityId) REFERENCES MailingCity (ID) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT FK_Staff_MailingPostalCode FOREIGN KEY (PostalCodeId) REFERENCES MailingPostalCode (ID) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT FK_Staff_MailingState FOREIGN KEY (StateId) REFERENCES MailingState (ID) ON DELETE SET NULL ON UPDATE CASCADE
);

-- Contact (generic contact records for students, staff, or other recipients)
CREATE TABLE `contact` (
  ID INT NOT NULL AUTO_INCREMENT,
  LastName VARCHAR(50) NOT NULL,
  FirstName VARCHAR(50) NULL,
  Email VARCHAR(200) NULL,
  Phone VARCHAR(30) NULL,
  Mobile VARCHAR(30) NULL,
  Note LONGTEXT NULL,
  PRIMARY KEY (ID)
);

-- StaffType
CREATE TABLE StaffType (
  StaffTypeID INT NOT NULL AUTO_INCREMENT,
  StaffTypeName VARCHAR(50) NOT NULL,
  StaffTypeDescription VARCHAR(200) NULL,
  IsSystemDefined TINYINT(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (StaffTypeID),
  UNIQUE KEY UIX_StaffType_StaffTypeName (StaffTypeName)
);

-- StaffStaffType (StaffID = Staff.ID)
CREATE TABLE StaffStaffType (
  ID INT NOT NULL AUTO_INCREMENT,
  StaffID INT NOT NULL,
  StaffTypeID INT NOT NULL,
  PrimaryFlag TINYINT(1) NULL DEFAULT 0,
  PRIMARY KEY (ID),
  UNIQUE KEY UQ_StaffStaffType_Staff_StaffType (StaffID, StaffTypeID),
  KEY IX_StaffStaffType_StaffTypeID (StaffTypeID),
  CONSTRAINT FK_StaffStaffType_Staff FOREIGN KEY (StaffID) REFERENCES Staff (ID) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT FK_StaffStaffType_StaffType FOREIGN KEY (StaffTypeID) REFERENCES StaffType (StaffTypeID)
);

-- Student (slim: identity, note, optional disability code, address refs, grade/school)
CREATE TABLE Student (
  ID INT NOT NULL AUTO_INCREMENT,
  Last_Name VARCHAR(20) NULL,
  First_Name VARCHAR(15) NULL,
  Middle_Name VARCHAR(50) NULL,
  Note LONGTEXT NULL,
  DisabilityCodeId INT NULL,
  Dob DATE NULL,
  Entry_Date DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  Grade INT NULL,
  School VARCHAR(10) NULL,
  Gender SMALLINT NULL,
  Email VARCHAR(250) NULL,
  CityId INT NULL,
  PostalCodeId INT NULL,
  StateId INT NULL,
  PRIMARY KEY (ID),
  KEY IX_Student_School (School),
  KEY IX_Student_Entry_Date (Entry_Date),
  CONSTRAINT FK_Student_DisabilityCode FOREIGN KEY (DisabilityCodeId) REFERENCES Disability_Codes (ID) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT FK_Student_MailingCity FOREIGN KEY (CityId) REFERENCES MailingCity (ID) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT FK_Student_MailingPostalCode FOREIGN KEY (PostalCodeId) REFERENCES MailingPostalCode (ID) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT FK_Student_MailingState FOREIGN KEY (StateId) REFERENCES MailingState (ID) ON DELETE SET NULL ON UPDATE CASCADE
);

-- StudentSchedule
CREATE TABLE StudentSchedule (
  ID INT NOT NULL AUTO_INCREMENT,
  `Sequence` INT NOT NULL,
  PreviousScheduleID INT NULL,
  StudentID INT NOT NULL,
  DBID INT NOT NULL,
  TripId INT NOT NULL,
  PUStopId INT NOT NULL,
  DOStopId INT NOT NULL,
  StudentRequirementId INT NULL,
  CrossStatus TINYINT(1) NULL,
  LastUpdated DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  WalkToStopDistance DECIMAL(20,6) NULL,
  LastUpdatedID INT NULL DEFAULT -999,
  LastUpdatedType SMALLINT NULL DEFAULT 1,
  SessionId TINYINT NULL,
  StopCrosser TINYINT(1) NULL,
  DistanceOnVehicle DECIMAL(38,8) NULL,
  PRIMARY KEY (ID),
  KEY IX_StudentSchedule_DBID_DoStopId_TripID (DBID, DOStopId, TripId),
  KEY IX_StudentSchedule_DBID_PuStopId_TripID (DBID, PUStopId, TripId),
  KEY IX_StudentSchedule_DBID_StudentID (DBID, StudentID),
  KEY IX_StudentSchedule_DBID_TripId_StudentRequirementId (DBID, TripId, StudentRequirementId),
  KEY IX_StudentSchedule_LastUpdated (LastUpdated),
  KEY IX_StudentSchedule_PreviousScheduleID_DBID (PreviousScheduleID, DBID),
  KEY IX_StudentSchedule_StudentRequirementId (StudentRequirementId)
);

-- Trip (slim: Note, TripDays; DriverID -> Staff.ID, VehicleID -> Vehicle.ID)
CREATE TABLE Trip (
  ID INT NOT NULL AUTO_INCREMENT,
  Name VARCHAR(150) NOT NULL,
  TripDays TINYINT NULL DEFAULT 0,
  Distance REAL NULL DEFAULT 0,
  DriverID INT NULL DEFAULT 0,
  VehicleID INT NULL DEFAULT 0,
  Note LONGTEXT NULL,
  GUID VARCHAR(50) NULL,
  PRIMARY KEY (ID)
);

-- CustomAttribute (custom attributes / formerly UDF)
CREATE TABLE CustomAttribute (
  ID INT NOT NULL AUTO_INCREMENT,
  DataType SMALLINT NOT NULL,
  `Type` INT NOT NULL,
  DisplayName VARCHAR(30) NOT NULL,
  Description LONGTEXT NULL,
  PickList TINYINT(1) NOT NULL DEFAULT 0,
  PickListMultiSelect TINYINT(1) NOT NULL DEFAULT 0,
  FormatString VARCHAR(50) NULL,
  `MinValue` INT NULL,
  `MaxValue` INT NULL,
  MaxLength INT NULL,
  DefaultBoolean TINYINT(1) NULL,
  DefaultNumeric DECIMAL(19,9) NULL,
  DefaultDate DATE NULL,
  DefaultTime TIME NULL,
  DefaultDatetime DATETIME NULL,
  DefaultText VARCHAR(255) NULL,
  DefaultPhoneNumber VARCHAR(100) NULL,
  DefaultZipCode VARCHAR(100) NULL,
  DefaultMemo LONGTEXT NULL,
  NumberPrecision SMALLINT NULL,
  TrueDisplayName VARCHAR(30) NULL,
  FalseDisplayName VARCHAR(30) NULL,
  ExpiresAfter SMALLINT NULL,
  Enabled TINYINT(1) NOT NULL DEFAULT 1,
  SystemDefined TINYINT(1) NOT NULL DEFAULT 0,
  LastUpdated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  LastUpdatedID INT NOT NULL,
  LastUpdatedType SMALLINT NULL DEFAULT 0,
  Required TINYINT(1) NOT NULL DEFAULT 0,
  Guid VARCHAR(100) NULL,
  AttributeFlag INT NOT NULL DEFAULT 0,
  ExternalID LONGTEXT NULL,
  RelatedDataType INT NULL,
  RelatedFormID INT NULL,
  `Function` INT NULL,
  ValueFormat INT NULL,
  RelatedDataField VARCHAR(100) NULL,
  RelatedDataFilterID INT NULL,
  RelatedStaticDataFilter LONGTEXT NULL,
  DefaultCase VARCHAR(1000) NULL,
  CaseDetail LONGTEXT NULL,
  DefaultEmail VARCHAR(200) NULL,
  IncludeCommas TINYINT(1) NULL,
  DefaultHyperlink VARCHAR(512) NULL,
  DefaultImage VARCHAR(1024) NULL,
  DefaultCaption VARCHAR(512) NULL,
  BoundarySetType SMALLINT NULL,
  BoundarySetName VARCHAR(1000) NULL,
  CreatedMethod TINYINT NULL,
  PRIMARY KEY (ID),
  UNIQUE KEY UIX_CustomAttribute_DataTypeDisplayName (DataType, DisplayName)
);

-- CustomAttributeType (attribute type lookup)
CREATE TABLE CustomAttributeType (
  ID INT NOT NULL AUTO_INCREMENT,
  `Type` VARCHAR(20) NOT NULL,
  PRIMARY KEY (ID)
);

-- Form (formerly UDGrid)
CREATE TABLE Form (
  ID INT NOT NULL AUTO_INCREMENT,
  DataType SMALLINT NOT NULL,
  Name VARCHAR(50) NOT NULL,
  Description VARCHAR(1000) NULL,
  Guid VARCHAR(100) NOT NULL,
  GridOptions LONGTEXT NULL,
  SystemDefined TINYINT(1) NOT NULL DEFAULT 0,
  LastUpdated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  AttributeFlag INT NOT NULL DEFAULT 0,
  ExternalID LONGTEXT NULL,
  IPAddressBoundary LONGTEXT NULL,
  CreatedOn DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `Public` TINYINT(1) NOT NULL DEFAULT 0,
  ExpiredOn DATETIME NULL,
  HasExpiredOn TINYINT(1) NOT NULL DEFAULT 0,
  GeofenceBoundaries LONGTEXT NULL,
  DisplayOneSection TINYINT(1) NOT NULL DEFAULT 0,
  ActiveOn DATETIME NULL,
  HasActiveOn TINYINT(1) NOT NULL DEFAULT 0,
  ThankYouMessage LONGTEXT NULL,
  AllowMultipleResponses TINYINT(1) NOT NULL DEFAULT 0,
  CreatedBy INT NULL DEFAULT 0,
  OneResponsePerRecipient TINYINT(1) NOT NULL DEFAULT 0,
  NotShowInFormfinder TINYINT(1) NOT NULL DEFAULT 0,
  AllowViewForm TINYINT(1) NOT NULL DEFAULT 0,
  AllowViewAllSubmittedForms TINYINT(1) NOT NULL DEFAULT 0,
  AllowSubmitNewForm TINYINT(1) NOT NULL DEFAULT 0,
  AllowRFIDScanning TINYINT(1) NOT NULL DEFAULT 0,
  SubmitFormOnRFIDScan TINYINT(1) NOT NULL DEFAULT 0,
  ShowCustomThankYouMessage TINYINT(1) NOT NULL DEFAULT 1,
  DisplayMessageLength INT NULL,
  FormLibraryId INT NULL,
  PRIMARY KEY (ID),
  UNIQUE KEY IX_Form_DataType_Name (DataType, Name),
  KEY IX_Form_Guid (Guid)
);

-- User
CREATE TABLE `User` (
  UserID INT NOT NULL AUTO_INCREMENT,
  Administrator TINYINT(1) NOT NULL DEFAULT 0,
  LoginID VARCHAR(50) NULL,
  Password VARBINARY(66) NULL,
  FirstName VARCHAR(50) NULL,
  LastName VARCHAR(50) NOT NULL,
  CreateDateTime DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  LastLoginDateTime DATETIME NULL,
  Email VARCHAR(200) NULL,
  Deactivated TINYINT(1) NOT NULL DEFAULT 0,
  DeactivatedDate DATETIME NULL,
  SecurityObjectId VARCHAR(50) NULL,
  SecurityUsername VARCHAR(50) NULL,
  PromptResetPassword TINYINT(1) NULL,
  Avatar LONGBLOB NULL,
  Phone VARCHAR(30) NULL,
  MFAPin VARCHAR(6) NULL,
  MFAPinCreateTime DATETIME NULL,
  MFALastLoginDateTime DATETIME NULL,
  WorkPhoneExt VARCHAR(4) NULL,
  CommunityUsername VARCHAR(255) NULL,
  CommunityPassword VARBINARY(255) NULL,
  Title VARCHAR(50) NULL,
  IsEmailVerified TINYINT(1) NULL,
  LastPasswordChangedDateTime DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  IsInvitationSent TINYINT(1) NOT NULL DEFAULT 1,
  PRIMARY KEY (UserID)
);

-- Vehicle (slim: Note, ModelInfo; Trip.VehicleID references Vehicle.ID)
CREATE TABLE Vehicle (
  ID INT NOT NULL AUTO_INCREMENT,
  Capacity INT NULL DEFAULT 0,
  Note LONGTEXT NULL,
  VehicleDetail LONGTEXT NULL,
  ContractorID INT NULL DEFAULT 0,
  VIN VARCHAR(60) NULL,
  ModelInfo VARCHAR(100) NULL,
  GPSID VARCHAR(100) NULL,
  AssetID VARCHAR(30) NULL,
  PRIMARY KEY (ID)
);

SET FOREIGN_KEY_CHECKS = 1;

-- Slim Vehicle: remove registration/insurance/dimensions/make-chassis fields; add VehicleDetail (JSON or plain text).
-- Run on databases created before this change. Fresh installs use migrations/001_schema_mysql.sql (Vehicle already slim).

ALTER TABLE Vehicle
  DROP COLUMN BrakeType,
  DROP COLUMN Cost,
  DROP COLUMN EmmissInsp,
  DROP COLUMN FuelCapacity,
  DROP COLUMN FuelType,
  DROP COLUMN Height,
  DROP COLUMN InspectionExp,
  DROP COLUMN InsuranceExp,
  DROP COLUMN InsuranceNum,
  DROP COLUMN `Length`,
  DROP COLUMN LicensePlate,
  DROP COLUMN LongName,
  DROP COLUMN MakeBody,
  DROP COLUMN MakeChassis,
  DROP COLUMN PurchaseDate,
  DROP COLUMN PurchasePrice,
  DROP COLUMN RegisExp,
  DROP COLUMN RegisNum,
  DROP COLUMN StateInspection,
  DROP COLUMN Width;

ALTER TABLE Vehicle
  ADD COLUMN VehicleDetail LONGTEXT NULL AFTER Comments;

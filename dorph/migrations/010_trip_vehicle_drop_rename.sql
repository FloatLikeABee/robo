-- Trip: remove audit, bus/aide flags, activity flag, description fields, disabled, weekday flags.
-- Vehicle: merge WC_Capacity into Capacity, drop wheelchair column; drop YearMade, BodyType, GUID.
--
-- After applying: restart/rebuild the Go API so handlers no longer SELECT Trip.Description.
-- Error 1054 Unknown column 'Description' means an old API process is still running.

ALTER TABLE Trip DROP INDEX IX_Trip_LastUpdated;

ALTER TABLE Trip
  DROP COLUMN LastUpdated,
  DROP COLUMN BusAide,
  DROP COLUMN ActivityTrip,
  DROP COLUMN Description,
  DROP COLUMN iDescription,
  DROP COLUMN Disabled,
  DROP COLUMN Monday,
  DROP COLUMN Tuesday,
  DROP COLUMN Wednesday,
  DROP COLUMN Thursday,
  DROP COLUMN Friday,
  DROP COLUMN Saturday,
  DROP COLUMN Sunday;

-- Prefer non-zero WC_Capacity when merging into Capacity
UPDATE Vehicle
SET Capacity = COALESCE(NULLIF(WC_Capacity, 0), Capacity);

ALTER TABLE Vehicle
  DROP COLUMN WC_Capacity,
  DROP COLUMN YearMade,
  DROP COLUMN BodyType,
  DROP COLUMN IF EXISTS GUID;

-- facility: JSON location (same shape as Activity.location — description + lat/lng)
SET @db := DATABASE();

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'facility' AND column_name = 'location');
SET @q := IF(@c = 0, 'ALTER TABLE `facility` ADD COLUMN location JSON NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

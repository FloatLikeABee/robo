-- Optional long-form description on main admin entities (500+ words).

SET @db := DATABASE();

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'member' AND column_name = 'description');
SET @q := IF(@c = 0, 'ALTER TABLE `member` ADD COLUMN description LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'facility' AND column_name = 'description');
SET @q := IF(@c = 0, 'ALTER TABLE `facility` ADD COLUMN description LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'employee' AND column_name = 'description');
SET @q := IF(@c = 0, 'ALTER TABLE `employee` ADD COLUMN description LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Activity' AND column_name = 'description');
SET @q := IF(@c = 0, 'ALTER TABLE Activity ADD COLUMN description LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'District' AND column_name = 'description');
SET @q := IF(@c = 0, 'ALTER TABLE District ADD COLUMN description LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'contact' AND column_name = 'description');
SET @q := IF(@c = 0, 'ALTER TABLE `contact` ADD COLUMN description LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

-- Widen Asset.description from VARCHAR(2000) when present (031).
SET @c := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = @db AND table_name = 'Asset' AND column_name = 'description');
SET @q := IF(@c > 0, 'ALTER TABLE Asset MODIFY COLUMN description LONGTEXT NULL', 'SELECT 1');
PREPARE p FROM @q; EXECUTE p; DEALLOCATE PREPARE p;

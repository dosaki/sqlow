ALTER TABLE activities ADD COLUMN migration_month VARCHAR(9);
UPDATE activities SET migration_month = 'September' where id = 1;
UPDATE activities SET migration_month = 'March' where id = 2;
ALTER TABLE activities ALTER COLUMN migration_month SET NOT NULL;
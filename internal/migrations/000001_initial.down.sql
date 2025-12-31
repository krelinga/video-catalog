-- Drop indices on links table
DROP INDEX IF EXISTS idx_links_work_uuid;
DROP INDEX IF EXISTS idx_links_source_uuid;

-- Drop links table
DROP TABLE IF EXISTS links;

-- Drop sources table
DROP TABLE IF EXISTS sources;

-- Drop works table
DROP TABLE IF EXISTS works;

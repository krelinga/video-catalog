-- Create works table
CREATE TABLE works (
    uuid UUID PRIMARY KEY,
    parent_uuid UUID REFERENCES works(uuid) ON DELETE CASCADE,
    kind VARCHAR NOT NULL CHECK (kind <> ''),
    body JSONB NOT NULL CHECK (body <> '{}'::jsonb)
);

-- Create sources table
CREATE TABLE sources (
    uuid UUID PRIMARY KEY,
    parent_uuid UUID REFERENCES sources(uuid) ON DELETE CASCADE,
    kind VARCHAR NOT NULL CHECK (kind <> ''),
    body JSONB NOT NULL CHECK (body <> '{}'::jsonb)
);

-- Create links table
CREATE TABLE links (
    source_uuid UUID NOT NULL REFERENCES sources(uuid) ON DELETE CASCADE,
    work_uuid UUID NOT NULL REFERENCES works(uuid) ON DELETE CASCADE,
    PRIMARY KEY (source_uuid, work_uuid)
);

-- Create indices on links table for efficient queries
CREATE INDEX idx_links_source_uuid ON links(source_uuid);
CREATE INDEX idx_links_work_uuid ON links(work_uuid);

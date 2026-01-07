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

-- Create plans table
CREATE TABLE plans (
    uuid UUID PRIMARY KEY,
    kind VARCHAR NOT NULL CHECK (kind <> ''),
    body JSONB NOT NULL CHECK (body <> '{}'::jsonb)
);

-- Create plan_inputs table
CREATE TABLE plan_inputs (
    plan_uuid UUID NOT NULL REFERENCES plans(uuid) ON DELETE CASCADE,
    source_uuid UUID NOT NULL REFERENCES sources(uuid) ON DELETE CASCADE,
    PRIMARY KEY (plan_uuid, source_uuid)
);

-- Create plan_outputs table
CREATE TABLE plan_outputs (
    plan_uuid UUID NOT NULL REFERENCES plans(uuid) ON DELETE CASCADE,
    work_uuid UUID NOT NULL REFERENCES works(uuid) ON DELETE CASCADE,
    PRIMARY KEY (plan_uuid, work_uuid)
);

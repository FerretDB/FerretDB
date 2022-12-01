DROP TABLE IF EXISTS oplog;

CREATE TABLE oplog (
    op varchar NOT NULL,
    wall timestamptz NOT NULL,
    ts bigint NOT NULL
);


CREATE OR REPLACE FUNCTION oplog_after_insert() RETURNS trigger AS $$

DECLARE
    r _ferretdb_internal.oplog%ROWTYPE;

BEGIN
    r.op := 'i';
    r.wall := now();
    r.ts := EXTRACT(epoch FROM r.wall);

    INSERT INTO _ferretdb_internal.oplog VALUES (r.*);
    RETURN NEW;
END;

$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION oplog_after_update() RETURNS trigger AS $$

DECLARE
    r _ferretdb_internal.oplog%ROWTYPE;

BEGIN
    r.op := 'i';
    r.wall := now();
    r.ts := EXTRACT(epoch FROM r.wall);

    INSERT INTO _ferretdb_internal.oplog VALUES (r.*);
    RETURN NEW;
END;

$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION oplog_after_delete() RETURNS trigger AS $$

DECLARE
    r _ferretdb_internal.oplog%ROWTYPE;

BEGIN
    r.op := 'i';
    r.wall := now();
    r.ts := EXTRACT(epoch FROM r.wall);

    INSERT INTO _ferretdb_internal.oplog VALUES (r.*);
    RETURN NEW;
END;

$$ LANGUAGE plpgsql;

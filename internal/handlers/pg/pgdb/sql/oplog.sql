CREATE SCHEMA IF NOT EXISTS _ferretdb_internal;

SET LOCAL search_path = _ferretdb_internal;

CREATE OR REPLACE TABLE oplog (
    table_name text,
    operation text,
    row_data jsonb
);

CREATE OR REPLACE FUNCTION oplog_insert_trigger_func() RETURNS trigger AS $$

DECLARE
    oplog_record oplog%ROWTYPE;

BEGIN
    SELECT nextval('oplog_id_seq') INTO oplog_record.id;
    oplog_record.table_name := TG_TABLE_NAME;
    oplog_record.operation := 'INSERT';
    oplog_record.row_data := row_to_json(NEW);
    INSERT INTO oplog VALUES (oplog_record);
    RETURN NEW;
END;

$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER oplog_insert_trigger
    AFTER INSERT ON oplog
    FOR EACH ROW EXECUTE PROCEDURE oplog_insert_trigger_func();

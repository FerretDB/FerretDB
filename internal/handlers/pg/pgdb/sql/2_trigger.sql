CREATE OR REPLACE TABLE oplog (
    ts bigint NOT NULL,
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

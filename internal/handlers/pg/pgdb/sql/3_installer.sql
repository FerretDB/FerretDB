CREATE OR REPLACE TRIGGER oplog_insert_trigger
    AFTER INSERT ON oplog
    FOR EACH ROW EXECUTE PROCEDURE oplog_insert_trigger_func();

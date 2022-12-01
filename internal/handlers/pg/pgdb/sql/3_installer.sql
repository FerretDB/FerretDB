CREATE OR REPLACE PROCEDURE oplog_install(table_name varchar) AS $$

BEGIN
    EXECUTE 'CREATE OR REPLACE TRIGGER oplog_after_insert ' ||
            'AFTER INSERT ON ' || quote_ident(table_name) || ' ' ||
            'FOR EACH ROW EXECUTE PROCEDURE _ferretdb_internal.oplog_after_insert()';

    EXECUTE 'CREATE OR REPLACE TRIGGER oplog_after_update ' ||
            'AFTER UPDATE ON ' || quote_ident(table_name) || ' ' ||
            'FOR EACH ROW EXECUTE PROCEDURE _ferretdb_internal.oplog_after_update()';

    EXECUTE 'CREATE OR REPLACE TRIGGER oplog_after_delete ' ||
            'AFTER DELETE ON ' || quote_ident(table_name) || ' ' ||
            'FOR EACH ROW EXECUTE PROCEDURE _ferretdb_internal.oplog_after_delete()';
END;

$$ LANGUAGE plpgsql;

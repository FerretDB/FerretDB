CREATE SCHEMA ferretdb_unifi;
CREATE TABLE ferretdb_unifi.format_version(version integer PRIMARY KEY);
INSERT INTO ferretdb_unifi.format_version VALUES (1);
CREATE TABLE ferretdb_unifi.capped_collections (
    collection_id bigint PRIMARY KEY REFERENCES documentdb_api_catalog.collections(collection_id) ON DELETE CASCADE,
    max_size bigint NOT NULL CHECK (max_size > 0),
    max_documents bigint NOT NULL DEFAULT 0 CHECK (max_documents >= 0),
    total_bytes bigint NOT NULL DEFAULT 0,
    total_documents bigint NOT NULL DEFAULT 0
);
CREATE TABLE ferretdb_unifi.capped_order (
    collection_id bigint NOT NULL REFERENCES ferretdb_unifi.capped_collections(collection_id) ON DELETE CASCADE,
    object_id documentdb_core.bson NOT NULL,
    position bigint GENERATED ALWAYS AS IDENTITY,
    document_size bigint NOT NULL,
    PRIMARY KEY(collection_id,object_id)
);
CREATE INDEX capped_order_position ON ferretdb_unifi.capped_order(collection_id,position);

CREATE FUNCTION ferretdb_unifi.enforce_capped() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog,documentdb_core,pg_temp AS $$
DECLARE
    identifier bigint := TG_ARGV[0]::bigint;
    cap ferretdb_unifi.capped_collections%ROWTYPE;
    previous_size bigint;
    current_size bigint;
    oldest documentdb_core.bson;
BEGIN
    SELECT * INTO cap FROM ferretdb_unifi.capped_collections WHERE collection_id=identifier FOR UPDATE;
    IF NOT FOUND THEN
        IF TG_OP='DELETE' THEN RETURN OLD; ELSE RETURN NEW; END IF;
    END IF;
    IF TG_OP='DELETE' THEN
        DELETE FROM ferretdb_unifi.capped_order
            WHERE collection_id=identifier AND object_id=OLD.object_id RETURNING document_size INTO previous_size;
        IF FOUND THEN
            UPDATE ferretdb_unifi.capped_collections
                SET total_bytes=total_bytes-previous_size,total_documents=total_documents-1 WHERE collection_id=identifier;
        END IF;
        RETURN OLD;
    END IF;
    current_size := octet_length(documentdb_core.bson_to_bytea(NEW.document));
    IF TG_OP='UPDATE' THEN
        SELECT document_size INTO previous_size FROM ferretdb_unifi.capped_order
            WHERE collection_id=identifier AND object_id=OLD.object_id;
        IF FOUND THEN
            UPDATE ferretdb_unifi.capped_order SET object_id=NEW.object_id,document_size=current_size
                WHERE collection_id=identifier AND object_id=OLD.object_id;
            UPDATE ferretdb_unifi.capped_collections SET total_bytes=total_bytes-previous_size+current_size
                WHERE collection_id=identifier;
            RETURN NEW;
        END IF;
    END IF;
    INSERT INTO ferretdb_unifi.capped_order(collection_id,object_id,document_size)
        VALUES(identifier,NEW.object_id,current_size);
    UPDATE ferretdb_unifi.capped_collections SET total_bytes=total_bytes+current_size,total_documents=total_documents+1
        WHERE collection_id=identifier;
    LOOP
        SELECT * INTO cap FROM ferretdb_unifi.capped_collections WHERE collection_id=identifier;
        EXIT WHEN cap.total_documents<=1 OR
            (cap.total_bytes<=cap.max_size AND (cap.max_documents=0 OR cap.total_documents<=cap.max_documents));
        SELECT object_id INTO oldest FROM ferretdb_unifi.capped_order
            WHERE collection_id=identifier ORDER BY position LIMIT 1;
        EXIT WHEN NOT FOUND;
        EXECUTE format('DELETE FROM %I.%I WHERE object_id=$1',TG_TABLE_SCHEMA,TG_TABLE_NAME) USING oldest;
    END LOOP;
    RETURN NEW;
END;
$$;

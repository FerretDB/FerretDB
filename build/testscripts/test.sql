-- Please do not merge changes in this file.
-- See https://pgtap.org/documentation.html#usingpgtap.

BEGIN;

SELECT plan(1);

SET documentdb_core.bsonUseEJson = on;

SELECT * FROM version();
SELECT * FROM documentdb_api.binary_extended_version();

SELECT * FROM documentdb_api.insert('test', '{"insert": "test", "documents": [{"v": 1}]}');

SELECT * FROM documentdb_api.db_stats('test', 1, true);

SELECT pass();
SELECT * FROM finish();

ROLLBACK;

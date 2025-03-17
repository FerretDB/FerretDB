-- Please do not merge changes in this file.

-- See https://pgtap.org/documentation.html#usingpgtap.

-- Start transaction and plan the tests.
BEGIN;
SELECT plan(1); -- number of test assertions

SET documentdb_core.bsonUseEJson = on;

-- Update the following example with your test.

SELECT pass();

-- Finish the tests and clean up.
SELECT * FROM finish();
ROLLBACK;

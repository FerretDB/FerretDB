// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.foo;
  t.drop();

  db.createCollection('foo', {capped: true, size: 10 * 1024 * 1024});

  t.insert({_id: 1});
  t.insert({_id: 2});

  const oneMinute = 1 * 60 * 1000;

  // defaults to 1000ms
  const defaultMaxTimeMS = 1 * 1000;

  let cmdRes = db.runCommand({
    find: 'foo',
    batchSize: 1,
    tailable: true,
    awaitData: true,
    maxTimeMS: oneMinute,
  });

  cmdRes = db.runCommand({
    getMore: cmdRes.cursor.id,
    collection: t.getName(),
    batchSize: 1,
  });

  // verify that the maxTimeMS value is not propagated to getMore
  const now = new Date();
  cmdRes = assert.commandWorked(
      db.runCommand({
        getMore: cmdRes.cursor.id,
        collection: t.getName(),
        batchSize: 1,
      }),
  );

  assert.eq(0, cmdRes.cursor.nextBatch.length);

  // allow the delta some margin of error
  assert.gte((new Date()) - now, defaultMaxTimeMS-100);

  print('test.js passed!');
})();

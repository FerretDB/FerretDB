// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.foo;
  t.drop();

  db.createCollection('foo', {capped: true, size: 1024 * 1024 * 10});

  t.insert({_id: 1});
  t.insert({_id: 2});

  const oneMinute = 1 * 60 * 1000;

  let cmdRes = db.runCommand({
    find: t.getName(),
    tailable: true,
    awaitData: true,
    batchSize: 1,
    maxTimeMS: oneMinute,
  });

  cmdRes = db.runCommand({
    getMore: cmdRes.cursor.id,
    collection: t.getName(),
    batchSize: 1,
    maxTimeMS: oneMinute,
  });
  assert.eq(2, cmdRes.cursor.nextBatch[0]._id);

  // should block until new data is available or timeout expires
  const now = new Date();
  cmdRes = assert.commandWorked(
      db.runCommand({
        getMore: cmdRes.cursor.id,
        collection: t.getName(),
        batchSize: 1,
        maxTimeMS: oneMinute,
      }),
  );
  // allow the delta some margin of error
  assert.gte((new Date()) - now, oneMinute-1000);

  print('test.js passed!');
})();

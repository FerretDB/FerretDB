// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.limit_reached;
  t.drop();

  t.insert({a: 1});
  t.insert({a: 2});
  t.insert({a: 3});
  t.insert({a: 4});

  let res = db.runCommand({find: t.getName(), batchSize: 3, limit: 4});
  assert.eq(res.cursor.firstBatch.length, 3);

  // assert that the cursor has been closed when the limit is reached
  res = db.runCommand({getMore: res.cursor.id, collection: t.getName(), batchSize: 1});
  assert.eq(1, res.cursor.nextBatch.length);
  assert.eq(0, res.cursor.id, 'cursor should be closed');

  print('test.js passed!');
})();

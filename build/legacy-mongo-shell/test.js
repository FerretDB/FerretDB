// Please do not merge changes in this file.

(function() {
  'use strict';

  DBQuery.shellBatchSize = 1;

  const coll = db.block_forever;
  coll.drop();

  // create a random index
  coll.createIndex({x: 1});

  coll.insert({});
  coll.insert({});
  
  // create a cursor
  const cursor = coll.find().batchSize(1);
  assert.eq(true, cursor.hasNext());
  cursor.next();

  print('about to block forever');
  assert.commandWorked(coll.dropIndex({x: 1}));

  print('test.js passed!');
})();

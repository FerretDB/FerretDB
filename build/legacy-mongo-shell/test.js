// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.batch_size;
  t.drop();

  // insert 102 documents
  const defaultBatchSize = 101;
  for (let i = 0; i < defaultBatchSize + 1; i++) {
    t.insert({_id: i});
  }

  const coll = 'batch_size';

  // assert that when no batchSize is specified, we use the default batchSize of 101
  var res = db.runCommand({find: coll, filter: {}});
  assert.eq(res.cursor.firstBatch.length, defaultBatchSize);

  // assert that batchSize works accordingly
  var res = db.runCommand({find: coll, filter: {}, batchSize: 1});
  assert.eq(res.cursor.firstBatch.length, 1);

  print('test.js passed!');
})();

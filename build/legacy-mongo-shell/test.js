// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.scan_capped_id;
  t.drop();

  const maxSize = 10000;
  const x = t.runCommand('create', {capped: true, size: maxSize});
  assert(x.ok);

  for (let i = 0; i < 100; i++) {
    t.insert({_id: i, x: 1});
  }

  const docSize = Object.bsonsize({_id: 0, x: 0});
  const totalSize = docSize * 100;
  assert.lt(totalSize, maxSize);

  const numDocsRemaining = (maxSize - totalSize) / docSize;

  // eslint-disable-next-line max-len
  print('number of documents that can be inserted before theoretical maximum size reached:', Math.floor(numDocsRemaining));

  for (let i = 100 - 1; i < numDocsRemaining + 100; i++) {
    t.insert({_id: i, x: 1});
  }

  const count = t.count();

  sleep(60 * 1000);

  assert.eq(t.count(), count, 'count should not change periodically');

  const totalCount = Math.floor(maxSize / docSize);
  t.insert({_id: totalCount, x: 1});
  assert.eq(null, t.findOne({_id: 0}), 'oldest entry should be overwritten');
  assert.eq(totalCount, t.count(), 'maximum size exceeded');

  print('test.js passed!');
})();

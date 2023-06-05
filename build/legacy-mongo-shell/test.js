// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.proj;
  const coll = 'proj';

  t.drop();

  const doc = {_id: 1, a: [1, 2, 3, 4]};

  assert.commandWorked(t.insert(doc));

  const filter = {'a': 1}; // positional operator needs a filter
  const proj = {'a.$': 1}; // add $ operator

  // eslint-disable-next-line max-len
  const res = t.runCommand({'find': coll, 'filter': filter, 'projection': proj});

  let expected = {'_id': 1, 'a': [1]};
  expected = [expected];
  assert.eq(res.cursor.firstBatch, expected);

  print('test.js passed!');
})();

// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.proj;
  const coll = 'proj';

  t.drop();

  const doc = {_id: 1, a: [{x: 1}, {x: 2}, {x: 3}]};

  assert.commandWorked(t.insert(doc));

  const filter = {'a.x': 1}; // positional operator needs a filter
  const proj = {'a.$': 1}; // add $ operator

  // eslint-disable-next-line max-len
  const res = t.runCommand({'find': coll, 'filter': filter, 'projection': proj});

  let expected = {'_id': 1, 'a': [{'x': 1}]};
  expected = [expected];
  assert.eq(res.cursor.firstBatch, expected);

  print('test.js passed!');
})();

// Move to test.js and run with `task mongo-test`.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  const init = [
    {_id: 'double', v: 42.13},
    {_id: 'double-whole', v: 42.0},
    {_id: 'double-zero', v: 0.0},
  ];

  coll.insertMany(init);

  const query = {v: {$gt: 42.0}};

  const expected = [
    {_id: 'double', v: 42.13},
  ];

  const actual = coll.find(query).toArray();
  assert.eq(expected, actual);

  print('test.js passed!');
})();

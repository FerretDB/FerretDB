// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  const init = [
    {_id: 1, arr: [1, 2, 3], foo: 42},
  ];

  coll.insertMany(init);

  const expected = [
    { _id: 1, arr: 1, foo: 42 },
    { _id: 1, arr: 2, foo: 42 },
    { _id: 1, arr: 3, foo: 42 }
  ];

  const actual = coll.aggregate([{$unwind: "$arr"}]).toArray();
  assert.eq(expected, actual);

  print('test.js passed!');
})();

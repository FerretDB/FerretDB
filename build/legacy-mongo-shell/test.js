// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  const double = 2305843009213694000.0;
  const int = NumberLong("2305843009213693952");

  // TODO:
  // negative
  // 1 << 53 f/l
  // 1 << 53 +1
  // 1 << 53 -1
  //
  // for 2<<60 (1<<61)
  // negative
  // 1 << 53 f/l
  // 1 << 53 +1
  // 1 << 53 -1
  const init = [
    {_id: 'double', v: double},
    {_id: 'int', v: int},
  ];

  coll.insertMany(init);

  let actual = coll.find().sort({v: 1, _id: 1}).toArray();
  let expected = init;

  assert.eq(expected, actual);

  actual = coll.find({v: double}).sort({v: 1, _id: 1}).toArray();
  expected = [
    {"_id": "double", "v": double},
    {"_id": "int", "v": int}
  ];
  assert.eq(expected, actual);

  print('test.js passed!');
})();

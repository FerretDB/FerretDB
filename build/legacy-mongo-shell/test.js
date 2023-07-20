// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  const init = [
    {_id: 'double', v: 42.13},
  ];

  coll.insertMany(init);


  const expected = [
    { _id: 42.13 },
  ];

  const actual = coll.coll.aggregate([ {$group: {_id: "$v"}}, {$project: {_id: "$_id"}}]).toArray();
  assert.eq(expected, actual);


  print('test.js passed!');
})();

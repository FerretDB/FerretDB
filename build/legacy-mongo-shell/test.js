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
    { _id: { foo: 'double', bar: 42.13 } },
  ];

  const actual = coll.aggregate([ {$group: {_id: {foo: "$_id", bar: "$v"}}}]).toArray();
  assert.eq(expected, actual);

  print('test.js passed!');
})();

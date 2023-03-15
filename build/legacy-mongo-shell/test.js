// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  const init = [
    {_id: 'array-double-desc', v: [40.0, 15.0, 10.0]},
    {_id: 'array-double-duplicate', v: [10.0, 10.0, 20.0]},
    {_id: 'array-double-empty', v: []},
  ];

  coll.insertMany(init);

  const pipeline = [
    {$group: {_id: "$v"}},
    {$sort: {_id: 1}},
  ];

  const expected = [
    {_id: []},
    {_id: [10.0, 10.0, 20.0]}, // same sort order as above since it uses the smallest value 10.0 for sorting.
    {_id: [40.0, 15.0, 10.0]}, // same sort order as below since it uses the smallest value 10.0 for sorting.
  ];

  const actual = coll.aggregate(pipeline).toArray();
  assert.eq(expected, actual);

  print('test.js passed!');
})();

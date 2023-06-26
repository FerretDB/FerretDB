// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.unwind;
  t.drop();

  t.insert({_id: 1, a: [1, 2, 3], b: 1});
  t.insert({_id: 2, a: [4, 5, 6], b: 2});

  const expected = [
    {
      '_id': 1,
      'a': 1,
      'b': 1,
    },
    {
      '_id': 1,
      'a': 2,
      'b': 1,
    },
    {
      '_id': 1,
      'a': 3,
      'b': 1,
    },
    {
      '_id': 2,
      'a': 4,
      'b': 2,
    },
    {
      '_id': 2,
      'a': 5,
      'b': 2,
    },
    {
      '_id': 2,
      'a': 6,
      'b': 2,
    },
  ];

  assert.eq(t.aggregate([{$unwind: '$a'}]).toArray(), expected);

  print('test.js passed!');
})();

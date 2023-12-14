// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  const filter = {"_id": 1, "f1.x": {$eq: 2}, "f1.y": {$ne: 3}}
  const update = {$set: {f2: 4}}
  const opts   = {upsert: true}

  coll.updateOne(filter, update, opts);

  // With `upsert: true`, filter fields without any operator (or)
  // fields with `$eq` operator should be added to the final document
  const expected = [
    {_id: 1, f1: {x: 2}, f2: 4},
  ];

  const actual = coll.find().toArray();

  assert.eq(expected, actual);

  print('test.js passed!');
})();
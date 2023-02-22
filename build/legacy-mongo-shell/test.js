// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.jstests_elemmatch_value;
  coll.drop();

  assert.commandWorked(coll.insert([
    {a: 5},
    {a: [5]}, // matches all the specified query criteria
    {a: [3, 7]},
  ]))

  assert.eq(coll.find({a: {$elemMatch: {$lt: 6, $gt: 4}}}, {_id: 0}).toArray(), [{a: [5]}]);
  print('test.js passed!');
})();

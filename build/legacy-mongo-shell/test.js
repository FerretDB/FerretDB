// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.foo;
  t.drop();

  t.save({_id: 1, a: 1, b: 1});
  t.save({_id: 2, a: 1, b: 1});

  const expected = [
    {_id: 1},
    {_id: 2},
  ];

  const proj = {_id: 1};

  assert.eq(expected, t.find({}, proj).toArray());

  // test that sorting doesn't break projection
  assert.eq(expected, t.find({}, proj).sort({_id: 1}).toArray());
  assert.eq(expected.reverse(), t.find({}, proj).sort({_id: -1}).toArray());

  print('test.js passed!');
})();

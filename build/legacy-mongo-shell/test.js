// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.foo;
  t.drop();

  t.save({a: [{}, {b: 1}]});
  assert.eq(1, t.find({'a.b': 1}).itcount()); // to show the expected behaviour.

  assert.eq(1, t.find({a: {$elemMatch: {b: 1}}}).itcount());

  print('test.js passed!');
})();

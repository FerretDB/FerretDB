// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.foo;
  t.drop();

  t.save({a: [{b: []}]});
  t.save({a: [{b: [{c: 1, d: 2}]}]});
  t.save({a: [{b: [{c: 3, d: 4}, {c: 1, d: 2}]}]});

  assert.eq(2, t.count({a: {$elemMatch: {b: {$elemMatch: {c: 1, d: 2}}}}}));

  print('test.js passed!');
})();

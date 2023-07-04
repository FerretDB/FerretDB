// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.a;
  t.drop();

  t.insert({a: []});
  t.insert({a: {b: 1}});
  t.insert({a: [{b: 1}]});
  t.insert({a: [{c: 1}, {b: 1}]});

  assert.eq(3, t.count({a: {b: 1}}));

  print('test.js passed!');
})();

// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.a;
  t.drop();

  t.insert({a: []});
  t.insert({a: [{b: 1}]});
  t.insert({a: [{c: 1}]});
  t.insert({a: [{b: 1}, {}]});

  assert.eq(2, t.count({'a.b': null}));
  assert.eq(2, t.count({a: {$elemMatch: {b: 1}}}));

  print('test.js passed!');
})();

// Move to test.js and run with `task mongo-test`.

(function() {
  'use strict';

  // version.txt: v0.7.1-29-gaf17fa7-dirty
  // tests that we can query an array of documents with dot notation.
  const t = db.dot_notation;
  t.drop();

  t.save({a: [{}, {b: 1}]});
  assert.eq(1, t.find({'a.b': 1}).itcount());

  print('test.js passed!');
})();

// Please do not merge changes in this file.

(function() {
  'use strict';

  // non-capped collection
  const t = db.foo;
  t.drop();

  t.insert({a: 1});
  t.insert({a: 2});
  t.insert({a: 3});
  t.insert({a: 4});

  let expected = [1, 2, 3, 4];

  let got = [];
  t.find({}).sort({$natural: 1}).forEach(function(d) {
    got.push(d.a);
  })

  assert.eq(expected, got, '$natural sort returned wrong order')

  print('test.js passed!');
})();

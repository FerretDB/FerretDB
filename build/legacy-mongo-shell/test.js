// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.pull;
  t.drop();

  let o = {_id: 1, a: []};
  for (let i = 0; i < 5; i++) {
    o.a.push({x: i, y: i});
  };

  t.insert(o);

  t.update({}, {$pull: {a: {x: 3}}});

  o.a = o.a.filter(function(z) {
    return z.x != 3;
  });

  assert.eq(o, t.findOne(), "element 3 should be pulled");

  print('test.js passed!');
})();

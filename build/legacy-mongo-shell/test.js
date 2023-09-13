// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.multi;
  t.drop();

  t.save({a: []});
  t.save({a: 1});
  t.save({a: []});

  assert.writeError(t.update({}, {$push: {a: 2}}, false, true));
  assert.eq(1, t.count({a: 2}));

  print('test.js passed!');
})();

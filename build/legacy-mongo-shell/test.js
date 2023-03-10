// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  const res = coll.update({'x.y': 2}, {$inc: {a: 7}}, true);
  assert.commandWorked(res);

  print('test.js passed!');
})();

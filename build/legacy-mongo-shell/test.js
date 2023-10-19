// Please do not merge changes in this file.

(function() {
  'use strict';

  // test uniqueness of _id
  const t = db.foo;
  t.drop();

  let res;

  res = t.save({_id: 3}); // A

  assert.commandWorked(res);

  res = t.insert({_id: 4, x: 99}); // B
  assert.commandWorked(res);

  // this should yield an error but we end up modifying A
  res = t.update({_id: 4}, {_id: 3, x: 99});
  assert.writeError(res);

  print('test.js passed!');
})();

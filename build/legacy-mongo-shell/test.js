// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.foo;
  t.drop();

  t.insert({_id: 1, a: 1});
  t.createIndex({a: 1});

  // a_1 already exists but command does not fail
  assert.commandWorked(t.createIndex({a: 1}));

  // _id_ already exists but command fails
  assert.commandWorked(t.createIndex({_id: 1}));

  print('test.js passed!');
})();

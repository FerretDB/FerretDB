// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.foo;
  t.drop();

  // ping works with the apiVersion set.
  assert.commandWorked(t.runCommand({ping: 1}));
  assert.commandWorked(t.runCommand({ping: 1, apiVersion: '1'}));
  jsTestLog("ping with apiVersion passed");

  assert.commandWorked(t.runCommand({insert: "foo", documents: [{}]}));
  assert.eq(1, t.count());

  // CRUD operations do not work.
  assert.commandWorked(t.runCommand({insert: "foo", documents: [{}], apiVersion: '1'}));
  assert.eq(2, t.count());
  jsTestLog("insert with apiVersion passed");

  print('test.js passed!');
})();

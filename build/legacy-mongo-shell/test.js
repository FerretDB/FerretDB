// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.allow_null;

  t.drop();
  t.insert({'a': 1});
  t.insert({'a': 2});
  t.insert({'a': 3});
  t.insert({'a': 4});

  const collName = 'allow_null';

  let res = assert.commandWorked(db.runCommand({distinct: collName, key: 'a'}));
  assert.eq([1, 2, 3, 4], res.values.sort());

  // eslint-disable-next-line max-len
  res = assert.commandWorked(db.runCommand({distinct: collName, key: 'a', query: {}}));
  assert.eq([1, 2, 3, 4], res.values.sort());

  // eslint-disable-next-line max-len
  res = assert.commandWorked(db.runCommand({distinct: collName, key: 'a', query: null}));
  assert.eq([1, 2, 3, 4], res.values.sort());

  print('test.js passed!');
})();

// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.type_miss_match;

  t.drop();
  t.insert({'a': 1});
  t.insert({'a': 1});
  t.insert({'a': 2});
  t.insert({'a': 3});

  // eslint-disable-next-line max-len
  assert.commandFailedWithCode(t.runCommand('distinct', {'key': {'a': 1}}), ErrorCodes.TypeMismatch);

  print('test.js passed!');
})();

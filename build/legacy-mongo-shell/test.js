// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.test;

  t.drop();

  const doc = {
    a: [{b: 1, c: []}]
  };

  assert.commandWorked(t.insert(doc));

  const res = t.update({}, {$inc: {'a.0.b': 1}, $push: {'a.0.c': 1}});
  assert.commandWorked(res);

  print('test.js passed!');
})();

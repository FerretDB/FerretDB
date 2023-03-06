// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.push_with_op;
  t.drop();

  const filter = {_id: 1, tags: {'$ne': 'a'}};

  assert.commandWorked(t.update(filter, {'$push': {tags: 'a'}}, true));
  assert.eq({_id: 1, tags: ['a']}, t.findOne(), 'A');

  print('test.js passed!');
})();

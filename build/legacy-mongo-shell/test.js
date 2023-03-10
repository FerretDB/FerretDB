// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  coll.insert({_id: 1});

  const expectedWriteError = coll.update({'_id': 1}, {$set: {'0..': 'val01'}});
  assert.eq(expectedWriteError.hasWriteError(), true);

  print('test.js passed!');
})();

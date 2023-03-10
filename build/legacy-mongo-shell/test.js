// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  const orig = {'_id': 1, 'c': 1, 'b': 2, 'a': 3};

  coll.update({_id: orig._id}, orig, {upsert: true});
  assert.eq(orig, coll.findOne());

  print('test.js passed!');
})();

// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  coll.findAndModify({query: {"non-existent": {$exists: false}}, upsert: true, update: {$set: {_id: "hello"}}});

  const query = {};

  const expected = [
    {_id: 'hello'},
  ];

  const actual = coll.find(query).toArray();
  assert.eq(expected, actual);

  print('test.js passed!');
})();

// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.foo;

  t.drop();

  const filter = {_id: 1};

  t.insert({_id: 1, text: 'a', n: 42});

  t.update(filter,
      {$set: {text: 'b'}, $setOnInsert: {_id: 2, n: 42}},
      {upsert: true},
  );

  assert.docEq([{_id: 1, text: 'b', n: 42}], t.find().toArray());

  t.update(filter,
      {$set: {n: -1}, $setOnInsert: {_id: 3, text: 'a'}},
      {upsert: true},
  );

  assert.docEq([{_id: 1, text: 'b', n: -1}], t.find().toArray());

  print('test.js passed!');
})();

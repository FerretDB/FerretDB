// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.field_path;
  t.drop();

  t.insert({_id: 1, a: 1});
  t.insert({_id: 2, a: 2});

  const expected = [{'_id': 1, 'a': 1}, {'_id': 2, 'a': 2}];

  const pipeline = {$project: {_id: 1, a: '$a'}};
  assert.eq(expected, t.aggregate(pipeline).toArray());

  print('test.js passed!');
})();

// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.nested_all;
  t.drop();

  t.save({_id: 1, a: [{x: 1}, {x: 2}]});
  t.save({_id: 2, a: [{x: 2}, {x: 3}]});
  t.save({_id: 3, a: [{x: 3}, {x: 4}]});

  assert.eq(1, t.find({'a.x': {$all: [1]}}).itcount(), 'A');
  assert.eq(2, t.find({'a.x': {$all: [2]}}).itcount(), 'B');

  // we should match the first document because it contains all the specified elements
  assert.eq(1, t.find({'a.x': {$all: [1, 2]}}).itcount(), 'C');

  print('test.js passed!');
})();

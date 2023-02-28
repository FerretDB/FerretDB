// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.pullall;
  t.drop();

  t.insert({_id: 1, a: [0, 2, 5, 5, 1, 0]});
  t.updateOne({_id: 1}, {$pullAll: {a: [0, 5]}});
  const res = t.findOne().a;
  res.sort();
  assert.eq([1, 2], res);

  t.drop();

  t.insert({a: [1, 2, 3]});
  t.update({}, {$pullAll: {a: [3]}});
  assert.eq([1, 2], t.findOne().a);
  t.update({}, {$pullAll: {a: [3]}});
  assert.eq([1, 2], t.findOne().a);

  t.drop();
  t.insert({a: [1, 2, 3]});
  t.update({}, {$pullAll: {a: [2, 3]}});
  assert.eq([1], t.findOne().a);
  t.update({}, {$pullAll: {a: []}});
  assert.eq([1], t.findOne().a);
  t.update({}, {$pullAll: {a: [5]}});
  assert.eq([1], t.findOne().a);
  t.update({}, {$pullAll: {a: [1, 5]}});
  assert.eq([], t.findOne().a);

  // test we can pull an embedded array
  t.drop();
  const ten = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
  t.insert({a: {b: ten}});
  t.update({}, {$pullAll: {'a.b': ten}});
  assert.eq([], t.findOne().a.b);


  t.drop();
  t.insert({m: 1});
  t.update({m: 1}, {$pullAll: {'a.b': [1]}});
  assert(('a' in t.findOne()) == false);
  t.update({m: 1}, {$pullAll: {'x.y': [1]}});
  assert(('z' in t.findOne()) == false);

  print('test.js passed!');
})();

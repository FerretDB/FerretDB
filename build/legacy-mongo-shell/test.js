/* eslint-disable require-jsdoc */
// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.not2_failures;

  function fail(query) {
    try {
      coll.find(query).itcount();
    } catch (e) {
      print(tojson(e));
      return;
    }
    assert.throws(() => coll.find(query).itcount(), [], query);
  }

  function doTest() {
    assert.commandWorked(coll.remove({}));

    assert.commandWorked(coll.insert({i: 'a'}));
    assert.commandWorked(coll.insert({i: 'b'}));

    // double negatives are not allowed
    fail({i: {$not: {$not: 'a'}}});

    fail({i: {$not: 'a'}});
    // unknown operator
    fail({i: {$not: {$ref: 'foo'}}});
    // $not cannot be empty
    fail({i: {$not: {}}});
    // $options needs a $regex
    fail({i: {$not: {$options: 'a'}}});
  }

  doTest();

  print('test.js passed!');
})();

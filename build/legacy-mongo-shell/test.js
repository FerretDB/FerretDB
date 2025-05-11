// Please do not merge changes in this file.

(function () {
  "use strict";

  db.coll.drop();

  db.coll.insertOne({a: 'a'.repeat(2000)});

  const actual = db.coll.findOneAndDelete({});

  assert.eq('a'.repeat(2000), actual.a);

  print("test.js passed!");
})();

// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

  const coll = db.test;

  coll.drop();

  const init = [
    { _id: "a", v: 1 },
    { _id: "b", v: 2 },
    { _id: "c", v: 3 }
  ];

  coll.insertMany(init);

  coll.find({}).limit(2).toArray();

  print("test.js passed!");
})();

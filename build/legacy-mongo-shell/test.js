// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

  const coll = db.test;

  coll.drop();

  const init = [
    { _id: "double", v: 42.13 },
    { _id: "double-whole", v: 42.0 },
    { _id: "double-zero", v: 0.0 },
  ];

  coll.insertMany(init);

  const expected = init;

  const res = db.test.runCommand({
    find: "test",
    filter: {
      $jsonSchema: {
		required: ["v"],
        bsonType: "object",
        properties: {
          v: {
            bsonType: "double",
          },
        },
      },
    },
    sort: {
      _id: 1,
    },
  });

  assert.isnull(res.errmsg);
  assert.eq(1,res.ok);

  const actual = res.cursor.firstBatch;

  assert.eq(expected, actual);

  print("test.js passed!");
})();

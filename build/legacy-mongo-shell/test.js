// Please do not merge changes in this file.

(function() {
  'use strict';

  const nameLong = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
  const nameShort = "test";

  db.getCollection(nameLong).drop();
  db.getCollection(nameShort).drop();

  db.getCollection(nameLong).insertOne({});

  assert.commandWorked(db.getCollection(nameLong).renameCollection(nameShort));

  assert.commandWorked(db.getCollection(nameShort).renameCollection(nameLong));

  print('test.js passed!');
})();

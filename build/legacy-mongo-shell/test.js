// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

  assert.commandWorked(db.test.createIndex({ v: 1 }, {name: "index"}));

  assert.commandWorked(db.test.dropIndex("index"));

  print("test.js passed!");
})();

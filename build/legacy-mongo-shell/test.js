// Please do not merge changes in this file.

(function () {
  "use strict";

  const actual = db.runCommand({ find: "test", "maxTimeMS": 0});
  assert.commandWorked(actual);

  print("test.js passed!");
})();

// Please do not merge changes in this file.

(function () {
  "use strict";

  const res = db.runCommand({ dropUser: "non-existent-user" });

  assert.commandFailedWithCode(res, 11);

  print("test.js passed!");
})();

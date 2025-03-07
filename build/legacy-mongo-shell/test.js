// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

	const res = db.runCommand({
		listDatabases: 1,
		filter: {"filter":{name:"nonexistent"}},
	});


  assert.eq([], res.databases);
  assert.eq(0,res.totalSize);

  print("test.js passed!");
})();

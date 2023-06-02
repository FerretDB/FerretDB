// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.match_div_zero;
  t.drop();

  // divisor cannot be 0, second argument is the divisor.
  let pipe = {$match: {a: {$mod: [0 /* invalid */, 0]}}};
  pipe = [pipe];

  // FailedToParse occurs when cursor is not present.
  const cmd = {pipeline: pipe, cursor: {batchSize: 0}};
  const res = t.runCommand('aggregate', cmd);

  const code = 2; // BadValue for $mod in a $match pipeline stage.

  const assertThrowsMsg = 'expected the following error code: ' + tojson(code);
  assert.commandFailedWithCode(res, code, assertThrowsMsg);

  print('test.js passed!');
})();

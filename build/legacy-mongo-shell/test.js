// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.collstats;
  t.drop();

  let pipeline = [{$match: {}}, {$collStats: {}}];
  const res = db.runCommand({aggregate: 'collstats', pipeline: pipeline, cursor: {}});
  assert.commandFailedWithCode(res, 40602);

  print('test.js passed!');
})();

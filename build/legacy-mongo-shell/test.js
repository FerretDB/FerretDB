// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.illegal_path_in_match;
  assert.commandWorked(coll.insert({a: 1}));

  const pipeline = [
    {$limit: 10}, // to prevent pushing the match into the query layer
    {$match: {'a.$c': 4}}, // legal path in query system, but illegal in aggregation
    // This inclusion-projection allows the planner to determine that the only necessary fields
    // we need to fetch from the document are "_id" (by default), "a.$c" (since we do a match
    // on it) and "dummy" since we include/rename it as part of this $project.

    // The reason we need to explicitly include a "dummy" field, rather than just including
    // "a.$c" is that, as mentioned before, a.$c is an illegal path in the aggregation system,
    // so if we use it as part of the project, the $project will fail to parse (and the
    // relevant code will not be exercised).
    {
      $project: {
        'newAndUnrelatedField': '$dummy',
      },
    },
  ];

  assert.commandFailedWithCode(db.runCommand({aggregate: 'illegal_path_in_match', pipeline: pipeline, cursor: {}}), 16410);

  print('test.js passed!');
})();

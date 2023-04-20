// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.proj;
  const coll = 'proj';

  t.drop();

  const verycomplexDoc = {
    _id: 1,
    a: {
      b: {
        c: {
          array: [
            {
              token: "aAOBX7fkiRB+XGH1oQ9fln7sM62ox06qzUKpaan7Bys="
            }
          ]
        }
      }
    }
  };

  assert.commandWorked(t.insert(verycomplexDoc));

  const filter = {'a.b.c.array.token': 'aAOBX7fkiRB+XGH1oQ9fln7sM62ox06qzUKpaan7Bys='};
  const proj = {'a.b.c.array.token': 1};

  assert.eq(verycomplexDoc._id, t.findOne(filter, proj)._id);

  const res = t.runCommand({'find': coll, 'filter': filter, 'projection': proj});

  // command must fail as we don't support $ projection
  // we also seem to project the _id field
  assert.commandFailed(res);

  print('test.js passed!');
})();

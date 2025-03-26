// Please do not merge changes in this file.

(function () {
  "use strict";

  const coll = db.test;

  coll.drop();

  const docsTotal = 1024;
  const docsPerBatch = 8;

  const batches = docsTotal / docsPerBatch;

  const v = Array.from({ length: 1024 * 1024 }, (_, i) => i);
  const batch = Array.from({ length: docsPerBatch }, () => ({ v }));

  for (let i = 0; i < batches; i++) {
    const start = new Date();
    coll.insertMany(batch);
    const duration = new Date() - start;
    print(`Inserted batch ${i + 1}/${batches} (${(i + 1) * docsPerBatch}/${docsTotal} documents) in ${duration}ms`);
  }

  shellPrint(db.adminCommand({ listDatabases: 1 }));

  print("test.js passed!");
})();

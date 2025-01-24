db.runCommand({
  aggregate: "collectionName",
  pipeline: [
    {
      $search: {
        cosmosSearch: {
          vector: "<vector>",
          path: "<path>",
          k: "<k>",

          // HNSW only
          efSearch: "<efSearch>",
        },
      },
    },
  ],
  cursor: {},
});

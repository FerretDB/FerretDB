db.runCommand({
  createIndexes: "<collectionName>",
  indexes: [
    {
      name: "<indexName>",
      key: {
        "<path>": "cosmosSearch",
      },
      cosmosSearchOptions: {
        kind: "<kind>",
        similarity: "<similarity>",
        dimensions: "<dimensions>",

        // HNSW only
        m: "<m>",
        efConstruction: "<efConstruction>",

        // IVF only
        numLists: "<numLists>",
      },
    },
  ],
});

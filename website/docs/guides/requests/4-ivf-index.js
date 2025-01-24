db.runCommand({
  createIndexes: "books",
  indexes: [
    {
      name: "vector_ivf_index",
      key: {
        vector: "cosmosSearch",
      },
      cosmosSearchOptions: {
        kind: "vector-ivf",
        similarity: "COS",
        dimensions: 12,
        numLists: 3,
      },
    },
  ],
});

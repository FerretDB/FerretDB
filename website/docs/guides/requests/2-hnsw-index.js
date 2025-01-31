db.runCommand({
  createIndexes: "books",
  indexes: [
    {
      name: "vector_hnsw_index",
      key: {
        vector: "cosmosSearch",
      },
      cosmosSearchOptions: {
        kind: "vector-hnsw",
        similarity: "COS",
        dimensions: 12,
        m: 16,
        efConstruction: 64,
      },
    },
  ],
});

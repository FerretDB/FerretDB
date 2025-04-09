db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      name: 'vector_hnsw_index',
      key: { vector: 'cosmosSearch' },
      cosmosSearchOptions: {
        kind: 'vector-hnsw',
        similarity: 'COS',
        dimensions: Int32(12),
        m: Int32(16),
        efConstruction: Int32(64)
      }
    }
  ],
  $db: '{{.Database}}'
})

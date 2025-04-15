db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      name: 'vector_ivf_index',
      key: { vector: 'cosmosSearch' },
      cosmosSearchOptions: { kind: 'vector-ivf', similarity: 'COS', dimensions: Int32(12), numLists: Int32(3) }
    }
  ],
  $db: '{{.Database}}'
})

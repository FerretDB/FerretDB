db.runCommand({
  dropIndexes: 'books',
  index: 'vector_hnsw_index'
})

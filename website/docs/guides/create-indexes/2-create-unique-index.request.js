db.runCommand({
  createIndexes: 'books',
  indexes: [{ key: { isbn: 1 }, name: 'unique_isbn_idx', unique: true }],
  $db: 'db'
})

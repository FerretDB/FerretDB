db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      key: { 'availability.format': 1 },
      name: 'ebook_india_idx',
      partialFilterExpression: {
        availability: {
          $elemMatch: {
            country: 'India',
            format: 'E-book'
          }
        }
      }
    }
  ]
})

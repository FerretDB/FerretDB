db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      key: {
        publisher: 1
      },
      name: 'publisher_recent_idx',
      partialFilterExpression: {
        'publication.year': {
          $gte: 2000
        }
      }
    }
  ]
})

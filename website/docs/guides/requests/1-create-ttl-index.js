db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      key: { 'publication.date': 1 },
      name: 'publication_date_ttl',
      expireAfterSeconds: 10
    }
  ]
})

db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      key: { 'publication.date': 1, 'analytics.average_rating': -1 },
      name: 'pub_date_rating_idx'
    }
  ]
})

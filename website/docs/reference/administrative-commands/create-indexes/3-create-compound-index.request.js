db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      key: {
        'analytics.average_rating': -1,
        'publication.date': 1
      },
      name: 'pub_date_rating_idx'
    }
  ]
})

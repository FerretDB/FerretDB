db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      key: {
        'reservation.date': 1
      },
      name: 'reservation_ttl',
      expireAfterSeconds: 60
    }
  ]
})

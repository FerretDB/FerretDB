db.runCommand({
  createIndexes: 'books',
  indexes: [{ key: { 'reservation.date': Int32(1) }, name: 'reservation_ttl', expireAfterSeconds: Int32(60) }],
  $db: '{{.Database}}'
})

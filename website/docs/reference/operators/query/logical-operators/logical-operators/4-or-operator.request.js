db.runCommand({
  find: 'books',
  filter: {
    $or: [
      {
        'authors.nationality': 'British'
      },
      {
        rating: {
          $gte: 4.5
        }
      }
    ]
  }
})

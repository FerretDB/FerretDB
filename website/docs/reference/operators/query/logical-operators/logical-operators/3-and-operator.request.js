db.runCommand({
  find: 'books',
  filter: {
    $and: [
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

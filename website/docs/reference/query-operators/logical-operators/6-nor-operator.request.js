db.runCommand({
  find: 'books',
  filter: {
    $nor: [
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

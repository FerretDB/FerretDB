response = {
  cursor: {
    id: Long(0),
    ns: 'db.books',
    firstBatch: [
      {
        _id: 'pride_prejudice_1813',
        title: 'Pride and Prejudice',
        authors: [
          {
            name: 'Jane Austen',
            nationality: 'British'
          }
        ],
        genres: ['Romance', 'Classic', 'Historical Fiction'],
        rating: 4.5
      }
    ]
  },
  ok: Double(1)
}

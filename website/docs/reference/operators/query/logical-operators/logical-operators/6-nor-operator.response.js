response = {
  cursor: {
    id: Long(0),
    ns: 'db.books',
    firstBatch: [
      {
        _id: 'moby_dick_1851',
        title: 'Moby Dick',
        authors: [
          {
            name: 'Herman Melville',
            nationality: 'American'
          }
        ],
        genres: ['Adventure', 'Classic', 'Sea Story'],
        rating: 4.3
      }
    ]
  },
  ok: Double(1)
}

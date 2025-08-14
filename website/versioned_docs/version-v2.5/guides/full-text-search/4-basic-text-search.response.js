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
        summary:
          'This novel portrays the life and challenges of Elizabeth Bennet as she ' +
          'navigates societal expectations, class prejudice, and romance. The book ' +
          'explores the evolving relationship between Elizabeth and Mr. Darcy, shedding ' +
          'light on the virtues of understanding and self-awareness.'
      }
    ]
  },
  ok: Double(1)
}

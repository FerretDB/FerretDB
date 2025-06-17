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
        summary:
          'Ishmael recounts his journey aboard the whaling ship Pequod under the leadership ' +
          'of the obsessed Captain Ahab, who is obsessed with hunting the legendary ' +
          'white whale, Moby Dick. The novel delves into themes of human struggle ' +
          'against nature and the destructive power of obsession.',
        score: Double(3)
      }
    ]
  },
  ok: Double(1)
}

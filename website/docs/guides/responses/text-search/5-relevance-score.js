response = {
  cursor: {
    id: Long(0),
    ns: '{{.Database}}.books',
    firstBatch: [
      {
        _id: 'moby_dick_1851',
        title: 'Moby Dick',
        author: 'Herman Melville',
        summary:
          'The narrative follows Ishmael and his voyage aboard the whaling ship Pequod, commanded by Captain Ahab, who is obsessed with hunting the elusive white whale, Moby Dick.',
        score: Int32(3)
      }
    ]
  },
  ok: 1.0
}

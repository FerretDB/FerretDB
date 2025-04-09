db.runCommand({
  insert: 'books',
  documents: [
    {
      _id: 'pride_prejudice_1813',
      title: 'Pride and Prejudice',
      author: 'Jane Austen',
      summary:
        'The novel follows the story of Elizabeth Bennet, a spirited young woman navigating love, societal expectations, and family drama in 19th-century England.'
    },
    {
      _id: 'moby_dick_1851',
      title: 'Moby Dick',
      author: 'Herman Melville',
      summary:
        'The narrative follows Ishmael and his voyage aboard the whaling ship Pequod, commanded by Captain Ahab, who is obsessed with hunting the elusive white whale, Moby Dick.'
    },
    {
      _id: 'frankenstein_1818',
      title: 'Frankenstein',
      author: 'Mary Shelley',
      summary:
        'Victor Frankenstein, driven by an unquenchable thirst for knowledge, creates a living being, only to face tragic consequences as his creation turns monstrous.'
    }
  ],
  $db: '{{.Database}}'
})

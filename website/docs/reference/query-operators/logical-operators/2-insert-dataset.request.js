db.runCommand({
  insert: 'books',
  documents: [
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
    },
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
    },
    {
      _id: 'clean_code_2008',
      title: 'Clean Code: A Handbook of Agile Software Craftsmanship',
      authors: [
        {
          name: 'Robert C. Martin',
          nationality: 'American'
        }
      ],
      genres: ['Computer Science', 'Software Engineering', 'Programming'],
      rating: 4.7
    }
  ]
})

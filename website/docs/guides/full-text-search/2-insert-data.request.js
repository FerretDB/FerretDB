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
      summary:
        'This novel portrays the life and challenges of Elizabeth Bennet as she ' +
        'navigates societal expectations, class prejudice, and romance. The book ' +
        'explores the evolving relationship between Elizabeth and Mr. Darcy, shedding ' +
        'light on the virtues of understanding and self-awareness.'
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
      summary:
        'Ishmael recounts his journey aboard the whaling ship Pequod under the leadership ' +
        'of the obsessed Captain Ahab, who is obsessed with hunting the legendary ' +
        'white whale, Moby Dick. The novel delves into themes of human struggle ' +
        'against nature and the destructive power of obsession.'
    },
    {
      _id: 'frankenstein_1818',
      title: 'Frankenstein',
      authors: [
        {
          name: 'Mary Shelley',
          nationality: 'British'
        }
      ],
      summary:
        'Victor Frankenstein creates life through unorthodox scientific methods ' +
        'but is horrified by his creation. As his creature seeks acceptance, the ' +
        'novel explores themes of responsibility, alienation, and the ethical limits ' +
        'of scientific experimentation.'
    }
  ]
})

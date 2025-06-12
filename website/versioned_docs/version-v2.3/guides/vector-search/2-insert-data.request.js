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
        'light on the virtues of understanding and self-awareness.',
      vector: [
        0.014391838572919369, -0.07001544535160065, 0.03249300271272659, 0.017455201596021652, -0.012363946065306664,
        0.04970458894968033, 0.05334962531924248, -0.04171367362141609, -0.042840130627155304, 0.038735587149858475,
        -0.036975011229515076, 0.02225673384964466
      ]
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
        'against nature and the destructive power of obsession.',
      vector: [
        -0.0016038859030231833, 0.08863562345504761, 0.006037247832864523, 0.044850509613752365, -0.019985735416412354,
        -0.017665650695562363, 0.07435955852270126, 0.0025448515079915524, -0.08427142351865768, 0.07445722818374634,
        -0.02302693948149681, -0.0778273269534111
      ]
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
        'of scientific experimentation.',
      vector: [
        -0.010190412402153015, 0.049356549978256226, -0.012309172190725803, 0.10420369356870651, 0.010599562898278236,
        0.057357728481292725, 0.02385704033076763, 0.04186723381280899, 0.003379989881068468, 0.02957085147500038,
        -0.08477196842432022, -0.0017921233084052801
      ]
    }
  ]
})

db.runCommand({
  find: 'books',
  filter: {
    $text: {
      $search: 'hunt whales'
    }
  },
  projection: {
    title: 1,
    authors: 1,
    summary: 1,
    score: {
      $meta: 'textScore'
    }
  },
  sort: {
    score: {
      $meta: 'textScore'
    }
  }
})

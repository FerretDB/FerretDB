db.runCommand({
  find: 'books',
  filter: {
    $text: {
      $search: 'romance'
    }
  },
  projection: {
    title: 1,
    authors: 1,
    summary: 1
  }
})

db.runCommand({
  find: 'books',
  filter: { $text: { $search: 'hunt whales' } },
  projection: { title: 1, author: 1, summary: 1, score: { $meta: 'textScore' } },
  sort: { score: { $meta: 'textScore' } }
})

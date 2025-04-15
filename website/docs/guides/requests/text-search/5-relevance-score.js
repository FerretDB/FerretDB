db.runCommand({
  find: 'books',
  filter: { $text: { $search: 'hunt whales' } },
  projection: { title: Int32(1), author: Int32(1), summary: Int32(1), score: { $meta: 'textScore' } },
  sort: { score: { $meta: 'textScore' } },
  $db: '{{.Database}}'
})

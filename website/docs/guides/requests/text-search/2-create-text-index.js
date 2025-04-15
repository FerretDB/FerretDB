db.runCommand({
  createIndexes: 'books',
  indexes: [{ key: { summary: 'text' }, name: 'summary_text_index' }],
  $db: '{{.Database}}'
})

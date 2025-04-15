db.runCommand({
  createIndexes: 'books',
  indexes: [{ key: { title: 'text', summary: 'text' }, name: 'title_summary_text_index' }],
  $db: '{{.Database}}'
})

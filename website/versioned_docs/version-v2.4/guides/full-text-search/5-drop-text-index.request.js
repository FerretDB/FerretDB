db.runCommand({
  dropIndexes: 'books',
  index: 'summary_text_index'
})

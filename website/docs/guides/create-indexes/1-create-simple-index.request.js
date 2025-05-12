db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      key: {
        title: 1
      },
      name: 'title_index'
    }
  ]
})

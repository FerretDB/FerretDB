db.runCommand({
  insert: 'books',
  documents: [
    {
      title: 'The Great Gatsby',
      publication: {
        date: new Date()
      }
    }
  ]
})

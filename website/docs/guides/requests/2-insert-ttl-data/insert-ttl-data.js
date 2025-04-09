db.runCommand({
  insert: 'books',
  documents: [
    {
      title: 'The Great Gatsby',
      author: 'F. Scott Fitzgerald',
      reservation: { user: 'Ethan Smith', date: ISODate('2025-03-15T11:00:00Z') }
    }
  ],
  $db: '{{.Database}}'
})

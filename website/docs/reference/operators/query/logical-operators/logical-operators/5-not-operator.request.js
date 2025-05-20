db.runCommand({
  find: 'books',
  filter: {
    rating: {
      $not: {
        $lt: 4.5
      }
    }
  }
})

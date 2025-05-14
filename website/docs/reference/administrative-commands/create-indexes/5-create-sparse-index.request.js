db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      key: {
        price: 1
      },
      name: 'sparse_price_index',
      sparse: true
    }
  ]
})

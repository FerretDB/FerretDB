db.runCommand({
  createIndexes: '<collection_name>',
  indexes: [
    {
      key: {
        '<field1>': '<index_type>',
        '<field2>': '<index_type>'
      },
      name: '<index_name>',
      unique: '<boolean>',
      partialFilterExpression: {
        '<field>': { '<operator>': '<value>' }
      },
      sparse: '<boolean>',
      expireAfterSeconds: '<int32>',
      cosmosSearchOptions: '<document>'
    }
  ]
})

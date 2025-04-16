db.runCommand({
  aggregate: 'books',
  pipeline: [
    {
      $search: {
        cosmosSearch: {
          vector: [
            Double(0.030856),
            Double(0.038531),
            Double(0.00079),
            Double(0.065121),
            Double(0.009282),
            Double(-0.056783),
            Double(0.029057),
            Double(0.021638),
            Double(0.012258),
            Double(0.055316),
            Double(-0.009759),
            Double(0.06137)
          ],
          path: 'vector',
          k: Int32(2)
        },
        returnStoredSource: true
      }
    }
  ],
  cursor: {}
})

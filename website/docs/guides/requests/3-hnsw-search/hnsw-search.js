db.runCommand({
  aggregate: "books",
  pipeline: [
    {
      $search: {
        cosmosSearch: {
          vector: [
            Double(0.022329),
            Double(0.0685),
            Double(0.030828),
            Double(0.090323),
            Double(-0.02827),
            Double(-0.036312),
            Double(0.024303),
            Double(-0.05155),
            Double(-0.067377),
            Double(0.01102),
            Double(-0.013403),
            Double(-0.004793),
          ],
          path: "vector",
          k: Int32(2),
          efSearch: Int32(40),
        },
      },
    },
  ],
  cursor: {},
});

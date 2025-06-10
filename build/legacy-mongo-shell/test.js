// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

  const coll = db.test;

  coll.drop();

  const arr1 = new Array(16384).fill(1).map(x => (Math.random() >= .5) ? 1 : 0)
  const arr2 = new Array(16384).fill(1).map(x => (Math.random() >= .5) ? 1 : 0)
  
  const init = [
    {
      _id: ObjectId('684211003840ee692afcbc1c'),
      meta: {
        annotations: [
          {
            label_id: '9b260dd55ab61a7600000000',
          }
        ]
      },
      data: BinData(9, arr1.buffer)
    },
    {
      _id: ObjectId('684211003840ee692afcbc1d'),
      meta: {
        annotations: [
          {
            label_id: '9b260dd55ab61a7600000000',
          }
        ]
      },
      data: BinData(9, arr2.buffer)
    }
  ];

  coll.insertMany(init);

  const query = [
    {
      "$group": {
        "_id": "$meta.annotations.label_id"
      }
    },
    {
      "$lookup": {
        "from": "test",
        "let": {
          "class": "$_id"
        },
        "pipeline": [
          {
            "$match": {
              "$expr": {
                "$eq": [
                  "$meta.annotations.label_id",
                  "$$class"
                ]
              }
            }
          },
          {
            "$limit": 4
          },
          {
            "$project": {
              "_id": 1,
              "data": "$data"
            }
          }
        ],
        "as": "instance"
      }
    },
    {
      "$unwind": "$instance"
    }
  ];

  const expected = [
    {
      "_id": [
        "9b260dd55ab61a7600000000"
      ],
      "instance": {
        "_id": ObjectId("684211003840ee692afcbc1c"),
        data: BinData(9, arr1.buffer)
      }
    },
    {
      "_id": [
        "9b260dd55ab61a7600000000"
      ],
      "instance": {
        "_id": ObjectId("684211003840ee692afcbc1d"),
        data: BinData(9, arr2.buffer)
      }
    }
  ];

  const actual = coll.aggregate(query).toArray();
  assert.eq(expected, actual);

  print("test.js passed!");
})();

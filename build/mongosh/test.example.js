docs = [
    { _id: 'double', v: 42.13 },
    { _id: 'double-whole', v: 42 },
    { _id: 'double-zero', v: 0 },
    { _id: 'double-max', v: 1.7976931348623157e+308 },
    { _id: 'double-smallest', v: 1.7976931348623157e+308 },
    { _id: 'double-big', v: 1.1529215e+18 },
    { _id: 'double-null', v: null },
    ]

db.foo.insertMany(docs)

db.foo.find({ "v": { "$size": -1 } })

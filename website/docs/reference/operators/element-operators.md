---
sidebar_position: 4
---

# Element query operators

Element query operators return data based on the existence of a specified field or the data type of a particular value.

| Operator             | Description                                                  |
| -------------------- | ------------------------------------------------------------ |
| [`$exists`](#exists) | returns documents where a field exists or does not exist     |
| [`$type`](#type)     | matches document containing elements with the specified type |

For the examples in this section, insert the following documents into the `electronics` collection:

```js
db.electronics.insertMany([
  {
     product: "laptop",
     price: 1500,
     stock: 5,
     discount: true,
     specifications: [
        {
           processor: "Intel Core i7"
        },
        {
           memory: 16
        }
     ]
  },
  {
     product: "phone",
     price: 800,
     stock: 10,
     discount: true,
     specifications: [
        {
           brand: "Apple"
        },
        {
           model: "iPhone 12"
        }
     ]
  },
  {
     product: "tablet",
     price: 500,
     stock: 15,
     discount: true,
     specifications: [
        {
           brand: "Samsung"
        },
        {
           model: "Galaxy Tab S7"
        }
     ]
  },
  {
     product: "keyboard",
     price: 100,
     stock: 20
  },
  {
     product: "mouse",
     price: 50,
     stock: 25,
     discount: null,
     specifications: []
  },
  {
     product: "monitor",
     price: 250,
     stock: 30,
     discount: true,
     specifications: [
        {
           size: 27
        },
        {
           resolution: "4K"
        }
     ]
  },
  {
     product: "printer",
     price: 150,
     stock: 35,
     discount: false
  },
  {
     product: "scanner",
     price: 100,
     stock: 40,
     discount: true,
     specifications: [
        {
           type: "flatbed"
        }
     ]
  }
])
```

## $exists

*Syntax*: `{ <field>: { $exists: <value> } }`

The `$exists` operator returns documents where a field exists or does not exist.

:::tip
If the `<boolean>` value is `true`, the query returns documents where the specified field exists, even if the value is `null` or an empty array.
If the `<boolean>` value is `false`, the query returns documents where the specified field does not exist.
:::

**Example:** To find documents in the `electronics` collection where the `specifications` field exists, use the `$exists` operator in the following query statement:

```js
db.electronics.find({
  specifications: {
     $exists: true
  }
})
```

The output:

```js
[
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b8e"),
   product: 'laptop',
   price: 1500,
   stock: 5,
   discount: true,
   specifications: [ { processor: 'Intel Core i7' }, { memory: 16 } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b8f"),
   product: 'phone',
   price: 800,
   stock: 10,
   discount: true,
   specifications: [ { brand: 'Apple' }, { model: 'iPhone 12' } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b90"),
   product: 'tablet',
   price: 500,
   stock: 15,
   discount: true,
   specifications: [ { brand: 'Samsung' }, { model: 'Galaxy Tab S7' } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b92"),
   product: 'mouse',
   price: 50,
   stock: 25,
   discount: null,
   specifications: []
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b93"),
   product: 'monitor',
   price: 250,
   stock: 30,
   discount: true,
   specifications: [ { size: 27 }, { resolution: '4K' } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b95"),
   product: 'scanner',
   price: 100,
   stock: 40,
   discount: true,
   specifications: [ { type: 'flatbed' } ]
 }
]
```

In the above output, the query returns all documents where the `specifications` field exists, even when the `field` has an empty value.

**Example:** If you want to find documents where the `specifications` field exists and has a specific value, use the `$exists` operator in conjunction with other operators.
The following query returns all documents where the `specifications` field exists and its value is an array:

```js
db.electronics.find({
  specifications: {
     $exists: true,
     $type: "array"
  }
})
```

The output:

```js
[
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b8e"),
   product: 'laptop',
   price: 1500,
   stock: 5,
   discount: true,
   specifications: [ { processor: 'Intel Core i7' }, { memory: 16 } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b8f"),
   product: 'phone',
   price: 800,
   stock: 10,
   discount: true,
   specifications: [ { brand: 'Apple' }, { model: 'iPhone 12' } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b90"),
   product: 'tablet',
   price: 500,
   stock: 15,
   discount: true,
   specifications: [ { brand: 'Samsung' }, { model: 'Galaxy Tab S7' } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b92"),
   product: 'mouse',
   price: 50,
   stock: 25,
   discount: null,
   specifications: []
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b93"),
   product: 'monitor',
   price: 250,
   stock: 30,
   discount: true,
   specifications: [ { size: 27 }, { resolution: '4K' } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b95"),
   product: 'scanner',
   price: 100,
   stock: 40,
   discount: true,
   specifications: [ { type: 'flatbed' } ]
 }
]
```

## $type

*Syntax*: `{ <field>: { $type: <datatype> } }`

The `$type` operator returns documents where the value of a field is of the specified BSON type.
The `<datatype>` parameter can be the type code or alias of the particular data type.

The following table lists the available BSON type codes and their corresponding aliases:

| Type code | Type               | Alias     |
| --------- | ------------------ | --------- |
| 1         | Double             | double    |
| 2         | String             | string    |
| 3         | Object             | object    |
| 4         | Array              | array     |
| 5         | Binary data        | binData   |
| 7         | ObjectId           | objectId  |
| 8         | Boolean            | bool      |
| 9         | Date               | date      |
| 10        | Null               | null      |
| 11        | Regular expression | regex     |
| 16        | 32-bit integer     | int       |
| 17        | Timestamp          | timestamp |
| 18        | 64-bit integer     | long      |
| 19        | Decimal128         | decimal   |
| -1        | Min key            | minKey    |
| 127       | Max key            | maxKey    |
| -128      | Number             | number    |

:::caution
`Decimal128`, `Min Key`, and `Max Key` are not currently implemented.
FerretDB supports the alias `number` which matches the following BSON types: `Double`, `32-bit integer`, and `64-bit integer` type values.
:::

:::info
FerretDB supports the alias `number` which matches the following BSON types: `Double`, `32-bit integer`, and `64-bit integer` type values.
:::

**Example:** The following operation query returns all documents in the `electronics` collection where the `discount` field has a boolean data type, which can be represented with the data code `8`:

```js
db.electronics.find({
  discount: {
     $type: 8
  }
})
```

This query can also be written using the alias of the specified data type.

```js
db.electronics.find({
  discount: {
     $type: "bool"
  }
})
```

The output:

```js
[
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b8e"),
   product: 'laptop',
   price: 1500,
   stock: 5,
   discount: true,
   specifications: [ { processor: 'Intel Core i7' }, { memory: 16 } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b8f"),
   product: 'phone',
   price: 800,
   stock: 10,
   discount: true,
   specifications: [ { brand: 'Apple' }, { model: 'iPhone 12' } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b90"),
   product: 'tablet',
   price: 500,
   stock: 15,
   discount: true,
   specifications: [ { brand: 'Samsung' }, { model: 'Galaxy Tab S7' } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b93"),
   product: 'monitor',
   price: 250,
   stock: 30,
   discount: true,
   specifications: [ { size: 27 }, { resolution: '4K' } ]
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b94"),
   product: 'printer',
   price: 150,
   stock: 35,
   discount: false
 },
 {
   _id: ObjectId("63a32fc7cf72d6203bb45b95"),
   product: 'scanner',
   price: 100,
   stock: 40,
   discount: true,
   specifications: [ { type: 'flatbed' } ]
 }
]
```

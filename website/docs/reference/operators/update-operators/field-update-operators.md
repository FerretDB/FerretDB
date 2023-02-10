---
sidebar_position: 1
---

# Field update operators

Field query operators allow you to specify a condition that must be met by documents in a collection in order to be returned in the result set. These operators can be used to perform various operations such as modification, comparison, and evaluation of fields in a document.

| Operator                       | Description                                                                             |
| ------------------------------ | --------------------------------------------------------------------------------------- |
| [`$currentDate`](#currentdate) | Sets the value of a field to current date, either as a Date or a Timestamp.             |
| [`$inc`](#inc)                 | Increments the value of the field by the specified amount.                              |
| [`$min`](#min)                 | Only updates the field if the specified value is less than the existing field value.    |
| [`$max`](#max)                 | Only updates the field if the specified value is greater than the existing field value. |
| [`$mul`](#mul)                 | Multiplies the value of the field by the specified amount.                              |
| [`$rename`](#rename)           | Renames a field.                                                                        |
| [`$set`](#set)                 | Sets the value of a field in a document.                                                |
| [`$setOnInsert`](#setoninsert) | Sets the value of a field if an update results in an insert of a document.              |
| [`$unset`](#unset)             | Removes the specified field from a document.                                            |

For the examples in this section, insert the following documents into the `catalog` collection:

```js
db.catalog.insertMany([
   {
      product: "bottle",
      price: 15,
      stock: 1,
      discount: true,
      variation:[
         {
            size:[
               "small",
               "medium",
               "large"
            ]
         },
         {
            color:[
               "black",
               "silver"
            ]
         }
      ]
   },
   {
      product: "spoon",
      price: 500,
      stock: 0,
      discount: true,
      variation:[
         {
            size:[
               "small",
               "medium",
               "large"
            ]
         },
         {
            color:[
               "silver",
               "white"
            ]
         }
      ]
   },
   {
      product: "cup",
      price: 100,
      stock: 14,
      discount: true,
      variation:[
         {
            size:[
               "small",
               "medium",
               "large"
            ]
         },
         {
            color:[
               "red",
               "black",
               "white"
            ]
         }
      ]
   },
   {
      product: "bowl",
      price: 56,
      stock: 5,
      discount: false,
      variation:[
         {
            size:[
               "small",
               "medium",
               "large"
            ]
         },
         {
            color:[
               "pink",
               "white",
               "red"
            ]
         }
      ]
   },

])
```

## $currentDate

*Syntax*: `{ <field>: { $currentDate: { <typeSpecification> } } }`

Syntax: `{ <field>: { $currentDate: true/false } }`

The `$currentDate` operator sets the value of a field to to either a Date object or a timestamp representing the current date.

For example, the following query sets the value of the `lastModified` field to the current date for all documents in the `catalog` collection:

```js
db.catalog.updateMany({}, { $currentDate: { lastModified: true } })
```

The output:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 8,
  modifiedCount: 8,
  upsertedCount: 0
}
```

## $inc

*Syntax*: `{ <field>: { $inc: { <amount> } } }`

The `$inc` operator increments the value of the field by the specified amount.

For example, the following query increments the value of the `stock` field by 1 for all documents in the `catalog` collection:

```js
db.catalog.updateMany({}, { $inc: { stock: 1 } })
```

The output:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 8,
  modifiedCount: 8,
  upsertedCount: 0
}
```

## $min

*Syntax*: `{ <field>: { $min: { <value> } } }`

The `$min` operator only updates the field if the specified value is less than the existing field value.

For example, the following query updates the value of the `stock` field to 0 if the value of the `stock` field is less than 0 for all documents in the `catalog` collection:

```js
db.catalog.updateMany({}, { $min: { stock: 0 } })
```

The output:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 8,
  modifiedCount: 8,
  upsertedCount: 0
}
```

## $max

*Syntax*: `{ <field>: { $max: { <value> } } }`

The `$max` operator only updates the field if the specified value is greater than the existing field value.

For example, the following query updates the value of the `stock` field to 0 if the value of the `stock` field is greater than 0 for all documents in the `catalog` collection:

```js
db.catalog.updateMany({}, { $max: { stock: 0 } })
```

The output:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 8,
  modifiedCount: 8,
  upsertedCount: 0
}
```

## $mul

*Syntax*: `{ <field>: { $mul: { <value> } } }`

The `$mul` operator multiplies the value of the field by the specified amount.

For example, the following query multiplies the value of the `stock` field by 2 for all documents in the `catalog` collection:

```js
db.catalog.updateMany({}, { $mul: { stock: 2 } })
```

The output:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 8,
  modifiedCount: 8,
  upsertedCount: 0
}
```

## $rename

*Syntax*: `{ $rename: { <field1>: <newName1>, ... } }`

The `$rename` operator renames a field.

For example, the following query renames the `product` field to `name` for all documents in the `catalog` collection:

```js
db.catalog.updateMany({}, { $rename: { product: "name" } })
```

The output:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 8,
  modifiedCount: 8,
  upsertedCount: 0
}
```

## $set

*Syntax*: `{ $set: { <field1>: <value1>, ... } }`

The `$set` operator sets the value of a field in a document.

For example, the following query sets the value of the `price` field to 100 for all documents in the `catalog` collection:

```js
db.catalog.updateMany({}, { $set: { price: 100 } })
```

The output:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 8,
  modifiedCount: 8,
  upsertedCount: 0
}
```

## $setOnInsert

*Syntax*: `{ $setOnInsert: { <field1>: <value1>, ... } }`

The `$setOnInsert` operator sets the value of a field if an update results in an insert of a document. Has no effect on update operations that modify existing documents.

For example, the following query sets the value of the `price` field to 100 for all documents in the `catalog` collection:

```js

db.catalog.updateMany({}, { $setOnInsert: { price: 100 } })
```

The output:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 8,
  modifiedCount: 8,
  upsertedCount: 0
}
```

## $unset

*Syntax*: `{ $unset: { <field1>: "", ... } }`

The `$unset` operator removes the specified field from a document.

For example, the following query removes the `price` field from all documents in the `catalog` collection:

```js

db.catalog.updateMany({}, { $unset: { price: "" } })
```

The output:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 8,
  modifiedCount: 8,
  upsertedCount: 0
}
```

:::note
Using `$unset` does not affect the indexes in your collections, so you may need to manually remove any unused indexes to reduce memory usage.
:::
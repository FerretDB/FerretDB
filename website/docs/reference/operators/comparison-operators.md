---
sidebar_position: 1
---

# Comparison query operators

Comparison query operators allow you to compare the elements in a document to a given query value.

Go to the comparison query operators:

| Operator       | Description                                                               |
| -------------- | ------------------------------------------------------------------------- |
| [`$eq`](#eq)   | Selects documents with elements that are equal to a given query value     |
| [`$gt`](#gt)   | Selects documents with elements that are greater than a given query value |
| [`$gte`](#gte) | Selects documents greater than or equal to specified query                |
| [`$lt`](#lt)   | Selects documents with elements that are less than a given query value    |
| [`$lte`](#lte) | Selects documents with elements less than or equal to given query value   |
| [`$in`](#in)   | Selects documents that contain the elements in a given array query        |
| [`$ne`](#ne)   | Selects documents with elements that are not equal to given query value   |
| [`$nin`](#nin) | Selects documents that do not contain the elements in a given array query |

For the examples in this section, insert the following documents into the `employees` collection:

```js
db.employees.insertMany([
  {
     name: {
        first: "Earl",
        last: "Thomas"
     },
     employeeID: 1234,
     age: 23,
     role: "salesperson",
     catalog: [
        "printer",
        "cardboard",
        "crayons",
        "books"
     ]
  },
  {
     name: {
        first: "Sam",
        last: "Johnson"
     },
     employeeID: 2234,
     age: 35,
     role: "salesperson",
     catalog: [
        "cabinet",
        "fridge",
        "blender",
        "utensils"
     ]
  },
  {
     name: {
        first: "Clarke",
        last: "Dane"
     },
     employeeID: 3234,
     age: 21,
     role: "salesperson",
     catalog: [
        "printer",
        "pencils",
        "crayons",
        "toys"
     ]
  }
])
```

## $eq

*Syntax*: `{ <field>: { $eq: <element> } }`

To select documents that exactly match a given query value, use the `$eq` operator.

This operator can be used to match values of different types, including documents, array, embedded documents, etc.

**Example:** The following operation queries the `employees` collection for all documents where the field `age` equals `21`.

```js
db.employees.find({
   age: {
      $eq: 21
   }
})
```

The response returns a single document that matches the query:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0e"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

**Example:** To query values in an embedded document, use [dot notation](../../understanding_ferretdb.md#dot-notation).
The following operation queries the `employees` collection for documents that match the field `first` in the embedded document `name`.

```js
db.employees.find({
   "name.first":{
      $eq: "Earl"
   }
})
```

The response returns a single document that matches the query:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  }
]
```

## $gt

*Syntax*: `{ <field>: { $gt: <element> } }`

To identify documents containing elements that have a greater value than the specified one in the query, use the `$gt` operator.

**Example:** Use the following operation to query for all the documents in the `employees` collection where the field `age` is greater than `21`.

```js
db.employees.find({
   age: {
      $gt: 21
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0d"),
    name: { first: 'Sam', last: 'Johnson' },
    employeeID: 2234,
    age: 35,
    role: 'salesperson',
    catalog: [ 'cabinet', 'fridge', 'blender', 'utensils' ]
  }
]
```

## $gte

*Syntax*: `{ <field>: { $gte: <element> } }`

Use the `$gte` to select document with elements that are greater than or equal to a specified value.

**Example:** The following operation selects documents based on the specified query, where the field `age` is greater than or equal to `21`.

```js
db.employees.find({
   age: {
      $gte: 21
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0d"),
    name: { first: 'Sam', last: 'Johnson' },
    employeeID: 2234,
    age: 35,
    role: 'salesperson',
    catalog: [ 'cabinet', 'fridge', 'blender', 'utensils' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0e"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

## $lt

*Syntax*: `{ field: { $lt: <element> } }`

Contrary to the `$gt` operator, the `$lt` operator is ideal for selecting documents with elements that are of a lesser value than that of the specified query.

**Example:** The following operation queries for documents where the field `age` is less than `25`.

```js
db.employees.find({
   age: {
      $lt: 25
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0e"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

## $lte

*Syntax*: `{ <field>: { $lte: <element> } }`

The `$lte` operator is the opposite of the `$gte` operator.
Use the `$lte` operator to select documents with elements that are less than or equal to the specified query value.

**Example:** The following operation queries for documents where the field `age` is less than or equal to `21`.

```js
db.employees.find({
   age: {
      $lte: 21
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0e"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

## $in

*Syntax*: `{ field: { $in: [<element1>, <element2>, ... <elementN> ] } }`

To select documents containing any of the listed elements in a specified array field, use the `$in` operator.

**Example:** The following operation queries the `employees` collection for documents where the value of the field `age` is either `21` or `35`.

```js
db.employees.find({
   age: {
      $in: [ 21, 35 ]
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0d"),
    name: { first: 'Sam', last: 'Johnson' },
    employeeID: 2234,
    age: 35,
    role: 'salesperson',
    catalog: [ 'cabinet', 'fridge', 'blender', 'utensils' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0e"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

## $ne

*Syntax*: `{ field: { $ne: <element> } }`

When selecting documents that do not match a specified query, use the `$ne` operator.

Use the `$ne` operator to select all the documents with elements that are not equal to a given query.

**Example:** The following operation queries the `employees` collection for documents where the field `age` is not equal to `21`.

```js
db.employees.find({
   age: {
      $ne: 21
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0d"),
    name: { first: 'Sam', last: 'Johnson' },
    employeeID: 2234,
    age: 35,
    role: 'salesperson',
    catalog: [ 'cabinet', 'fridge', 'blender', 'utensils' ]
  }
]
```

## $nin

*Syntax*: `{ <field>: { $nin: [ <element1>, <element2> ... <elementN> ] } }`

The `$nin` does exactly the opposite of the `$in` operator.
Use the `$nin` operator when selecting documents that do match or contain any of the elements listed in an array query.

**Example:** The following operation queries the `employees` collection for documents where the value of the field `age` is not `21` or `35`.

```js
db.employees.find({
   age: {
      $nin: [ 21, 35 ]
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  }
]

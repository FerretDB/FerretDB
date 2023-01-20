---
sidebar_position: 1
---

# Comparison query operators

Comparison query operators return data that matches specific query conditions.
Go to the comparison query operators:

| Operator       | Description                                                                     |
| -------------- | ------------------------------------------------------------------------------- |
| [`$eq`](#eq)   | Matches documents equal to specified query                                      |
| [`$gt`](#gt)   | Matches documents greater than specified query                                  |
| [`$gte`](#gte) | Matches documents greater than or equal to specified query                      |
| [`$lt`](#lt)   | Matches documents less than specified query                                     |
| [`$lte`](#lte) | Matches documents less than or equal to specified query                         |
| [`$in`](#in)   | Matches documents containing values in a specified array query                  |
| [`$ne`](#ne)   | Matches documents that are not equal to specified query                         |
| [`$nin`](#nin) | Matches documents that do not contain values present in a specified array query |

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

*Syntax*: `{ field: { $eq: value } }`

The equality operator `$eq` selects documents that matches the specified query.
The following operation queries the `employees` collection for all documents where the field `age` equals `21`.

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

To query values in an embedded document, use [dot notation](../../understanding_ferretdb.md#dot-notation).
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

*Syntax*: `{ field: { $gt: value } }`

The greater than operator `$gt` selects documents that are greater than the specified query.

Use the following operation to query for all the documents in the `employees` collection where the field `age` is greater than `21`.

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

*Syntax*: `{ field: { $gte: value } }`

The greater than or equal operator `$gte` selects document with values that are greater than or equal to the specified query.
The following operation selects documents based on the specified query, where the field `age` is greater than or equal to `21`.

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

*Syntax*: `{ field: { $lt: value } }`

The less than operator `$lt` selects document with values that are less than the specified query.
The following operation queries for documents where the field `age` is less than `25`.

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

*Syntax*: `{ field: { $lte: value } }`

The less than or equal operator `$lte` selects document with values that are less than or equal to the specified query.
The following operation queries for documents where the field `age` is less than or equal to `21`.

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

*Syntax*: `{ field: { $in: [<value1>, <value2>, ... <valueN> ] } }`

The `$in` operator selects documents that contain any of the values in a specified array.
For a document to be returned, the field must contain at least one of the values in the specified array query.

The following operation queries the `employees` collection for documents where the value of the field `age` is either `21` or `35`.

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

*Syntax*: `{ field: { $ne: value } }`

The `$ne` operator selects all the documents that are not equal to the specified query.
The following operation queries the `employees` collection for documents where the field `age` is not equal to `21`.

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

*Syntax*: `{ field: { $nin: [ <value1>, <value2> ... <valueN> ] } }`

The `$nin` operator selects documents that does not contain any of the values in a specified array.
The following operation queries the `employees` collection for documents where the value of the field `age` is not `21` or `35`.

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

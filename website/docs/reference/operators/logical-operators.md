---
sidebar_position: 2
---

# Logical query operators

Logical query operators return data based on specified query expressions that are either true or false.

| Operator       | Description                                                          |
| -------------- | -------------------------------------------------------------------- |
| [`$and`](#and) | Joins all query expressions with a logical AND operator              |
| [`$or`](#or)   | Joins all query expressions with a logical OR operator               |
| [`$not`](#not) | Returns all documents that do NOT match a query expression           |
| [`$nor`](#nor) | Returns all documents that do not match any of the query expressions |

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

## $and

*Syntax*: `{ $and: [ { <expression1> }, { <expression2> } , ... , { <expressionN> } ] }`

The `$and` operator joins one or more query expressions, and returns data that matches all the expressions.

Select documents that satisfy both of these expressions in the `catalog` collection:

* `price` field is less than `100` **AND**
* `stock` field is not `0`

```js
db.catalog.find({
   $and:[
      {
         price:{
            $lt: 100
         }
      },
      {
         stock:{
            $ne: 0
         }
      }
   ]
})
```

The output:

```js
[
  {
    _id: ObjectId("639ba4a0071b6bed396a8f13"),
    product: 'bottle',
    price: 15,
    stock: 1,
    discount: true,
    variation: [
      { size: [ 'small', 'medium', 'large' ] },
      { color: [ 'black', 'silver' ] }
    ]
  },
  {
    _id: ObjectId("639ba4a0071b6bed396a8f16"),
    product: 'bowl',
    price: 56,
    stock: 5,
    discount: false,
    variation: [
      { size: [ 'small', 'medium', 'large' ] },
      { color: [ 'pink', 'white', 'red' ] }
    ]
  }
]
```

## $or

*Syntax*: `{ $or: [ { <expression1> }, { <expression2> } , ... , { <expressionN> } ] }`

The `$or` operator joins one or more query expressions, and returns data that matches at least one of the expressions.

Select the documents that match these expressions:

* `discount` field is `true` *and* `stock` field is not `0` **OR**
* `price` field is less than or equal to `60`

```js
db.catalog.find({
   $or:[
      {
         $and:[
            {
               discount: true
            },
            {
               stock:{
                  $ne: 0
               }
            }
         ]
      },
      {
         price:{
            $lte: 60
         }
      }
   ]
})
```

The output:

```js
[
  {
    _id: ObjectId("639ba4a0071b6bed396a8f13"),
    product: 'bottle',
    price: 15,
    stock: 1,
    discount: true,
    variation: [
      { size: [ 'small', 'medium', 'large' ] },
      { color: [ 'black', 'silver' ] }
    ]
  },
  {
    _id: ObjectId("639ba4a0071b6bed396a8f15"),
    product: 'cup',
    price: 100,
    stock: 14,
    discount: true,
    variation: [
      { size: [ 'small', 'medium', 'large' ] },
      { color: [ 'red', 'black', 'white' ] }
    ]
  },
  {
    _id: ObjectId("639ba4a0071b6bed396a8f16"),
    product: 'bowl',
    price: 56,
    stock: 5,
    discount: false,
    variation: [
      { size: [ 'small', 'medium', 'large' ] },
      { color: [ 'pink', 'white', 'red' ] }
    ]
  }
]
```

## $not

*Syntax*: `{ field: { $not: { <expression> } } }`

The `$not` operator selects documents that do not match the specified expression.

The following operation selects documents that do not satisfy the specified expression, where the `stock` field is not less than `5`.

```js
db.catalog.find({
    stock: {
        $not: {
            $lt: 5
        }
    }
})
```

The output:

```js
[
  {
    _id: ObjectId("639ba4a0071b6bed396a8f15"),
    product: 'cup',
    price: 100,
    stock: 14,
    discount: true,
    variation: [
      { size: [ 'small', 'medium', 'large' ] },
      { color: [ 'red', 'black', 'white' ] }
    ]
  },
  {
    _id: ObjectId("639ba4a0071b6bed396a8f16"),
    product: 'bowl',
    price: 56,
    stock: 5,
    discount: false,
    variation: [
      { size: [ 'small', 'medium', 'large' ] },
      { color: [ 'pink', 'white', 'red' ] }
    ]
  }
]
```

## $nor

*Syntax*: `{ $nor: [ { <expression1> }, { <expression2> }, ...  { <expressionN> } ] }`

The `$nor` operator selects documents that do not match any of the specified expressions.

Select the documents that fail to match any of these expressions:

* `discount` field is `true` *and* `stock` field is not `0`
* `price` field is less than or equal to `60`

```js
db.catalog.find({
   $nor:[
      {
         $and:[
            {
               discount: true
            },
            {
               stock:{
                  $ne: 0
               }
            }
         ]
      },
      {
         price:{
            $lte: 60
         }
      }
   ]
})
```

The output:

```js
[
  {
    _id: ObjectId("639ba4a0071b6bed396a8f14"),
    product: 'spoon',
    price: 500,
    stock: 0,
    discount: true,
    variation: [
      { size: [ 'small', 'medium', 'large' ] },
      { color: [ 'silver', 'white' ] }
    ]
  }
]
```

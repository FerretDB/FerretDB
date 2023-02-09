---
sidebar_position: 5
---

# Evaluation query operators

Evaluation query operators return data based on the evaluation of a specified expression.

| Operator           | Description                                                                                   |
| ------------------ | --------------------------------------------------------------------------------------------- |
| [`$mod`](#mod)     | Matches documents where the value of a field divided by a divisor has the specified remainder |
| [`$regex`](#regex) | Matches documents where a specified field matches a specified regular expression pattern      |

For the examples in this section, insert the following documents into the `catalog` collection:

```js
db.catalog.insertMany([
   {
      product: "bottle",
      price: 15,
      stock: 1,
   },
   {
      product: "spoonn",
      price: 500,
      stock: 0,
   },
   {
      product: "cup",
      price: 100,
      stock: 14,
   },
   {
      product: "BoWL",
      price: 56,
      stock: 5,
   },
   {
      product: "boTtLe",
      price: 20,
      stock: 3,
   }
])
```

## $mod

*Syntax*: `{ <field>: { $mod: [ <divisor>, <remainder> ] } }`

The `$mod` operator matches documents where the value of a field divided by a specified divisor has a specified remainder (i.e. the field value corresponds to `remainder` modulo `divisor`).

For example, the following query returns all the documents where the value of the "stock" field is evenly divisible by 2:

```js
db.catalog.find({
   stock:{
      $mod: [ 2, 0 ]
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("63e3ac0184f488929a3f737a"),
    product: 'spoon',
    price: 500,
    stock: 0
  },
  {
    _id: ObjectId("63e3ac0184f488929a3f737b"),
    product: 'cup',
    price: 100,
    stock: 14
  }
]
```

:::caution
Note that the `$mod` expression returns an error if you only have a single element in the array, more than two elements in the array, or if the array is empty.
It also rounds down decimal input down to zero (e.g. `$mod: [ 3.5 , 2 ]` is executed as `$mod: [ 3 , 2 ]`).
:::

## $regex

*Syntax*: `{ <field>: { $regex: 'pattern', $options: '<options>' } }`

Other syntaxes: `{ <field>: { $regex: /pattern/, $options: '<options>' } }` and `{ <field>: /pattern/<options> }`.

The `$regex` operator matches documents where the value of a field matches a specified regular expression pattern.

The following query returns all the documents where the value of the "product" field starts with the letter "b":

```js
db.catalog.find({
   product:{
      $regex: /^b/
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("63e3ac0184f488929a3f7379"),
    product: 'bottle',
    price: 15,
    stock: 1
  }
]
```

`$options` is an optional parameter that specifies the regular expression options to use, such as:

* case-insensitivity (`i`)
* multi-line matching (`m`)
* dot character matching (`s`)

:::note
The regex option for ignoring white spaces (`x`) is not currently supported.
Follow [here](https://github.com/FerretDB/FerretDB/issues/592) for more updates.
:::

To perform case-insensitive matching, use the `i` option in the `regex` expression.
The following query returns all the documents where the value of the "product" field is equal to "bottle" (case-insensitive):

```js
db.catalog.find({
   product: {
      $regex: /bottle/i
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("63e3ac0184f488929a3f7379"),
    product: 'bottle',
    price: 15,
    stock: 1
  },
  {
    _id: ObjectId("63e3ac0184f488929a3f737d"),
    product: 'boTtLe',
    price: 20,
    stock: 3
  }
]
```

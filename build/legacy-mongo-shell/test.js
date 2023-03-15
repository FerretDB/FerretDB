// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  // As double's mantissa is 53bits long, we expect that every number that's <= 2<<53-1 can be represented precisely.
  //
  // Every Javscript Number type is stored as IEEE754 double floating point.
  let init = [
    {_id: 'double-max-prec', v: Number.MAX_SAFE_INTEGER}, // Double 2**53-1 is a last precise number = 9007199254740991.0
    {_id: 'double-max-prec-long', v: NumberLong(Number.MAX_SAFE_INTEGER)}, // Long 2**53-1 is precise but not last = 9007199254740991
  ];

  coll.insertMany(init);

  // 9007199254740991.0 == 9007199254740991.0; 9007199254740991.0 == 9007199254740991
  let actual = coll.find({v: Number.MAX_SAFE_INTEGER}).toArray();
  assert.eq(init, actual);

  // 9007199254740991 == 9007199254740991.0; 9007199254740991 == 9007199254740991
  actual = coll.find({v: NumberLong(Number.MAX_SAFE_INTEGER)}).toArray();
  assert.eq(init, actual);

  let bigNumbers = [
    {_id: 'double-big', v: 2 * Math.pow(2,60)}, // 2**60 double loses it's precision = 2305843009213694000.0
	{_id: 'double-big-long', v: NumberLong(2 * Math.pow(2,60))}, // 2**60 long doesn't loose precision = 2305843009213693952
  ];

  let bigNumbersDontMatch = [
	// I've tried to do `NumberLong(2*Math.pow(2,60))+NumberLong(1)` and other combinantions but it seems 
	// that everytime mongodb got 2305843009213694000 or Long("2305843009213693952"). Not sure why, but I assume it's because of javascript.
	{_id: 'double-big-long-plus', v: NumberLong("2305843009213693953")}, // 2**60 Long +1 = 2305843009213693953

  ];

  coll.insertMany( (new Array).concat(bigNumbers, bigNumbersDontMatch));

  // If one of numbers in comparison is "unsafe" double, both long and double values loose their precision, so:
  // 2305843009213694000.0 == 2305843009213693952
  // 2305843009213694000.0 == 2305843009213694000.0
  //
  // Technically speaking, here following statement should be true:
  // 2305843009213694000.0 == 2305843009213693953
  //
  // Although at this point test on mongodb fails. I don't know how is it possible that 
  // 2305843009213693952 is rounded to 2305843009213694000.0 correctly, but
  // 2305843009213693953 maybe not?
  actual = coll.find({v: 2 * Math.pow(2,60)}).toArray();
  assert.eq((new Array).concat(bigNumbers, bigNumbersDontMatch), actual);

  // The same logic applies to the long filter:
  // 2305843009213693952 == 2305843009213694000.0
  //
  // Although if both values are longs, they are compared as all, so:
  // 2305843009213693952 == 2305843009213693952
  // but...
  // 2305843009213693952 != 2305843009213693953
  actual = coll.find({v: NumberLong(2 * Math.pow(2,60))}).toArray();
  assert.eq(bigNumbers, actual);



  print('test.js passed!');
})();

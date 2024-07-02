Contributing
============

First of all, thank you for considering to contribute to Statsviz!

Pull-requests are welcome!


## Go library

The Statsviz Go public API is relatively light so there's not much to do and at
the moment it's unlikely that the API will change. However new options can be
added to `statsviz.Register` and `statsviz.NewServer` without breaking
compatibility.

That being said, there may be things to improve in the implementation, any
contribution is very welcome!

Big changes should be discussed on the issue tracker prior to start working on
the code.

If you've decided to contribute, thank you so much, please comment on the existing 
issue or create one stating what you want to tackle and why.


## User interface (html/css/javascript)

The user interface aims to be simple, light and minimal.

Assets are located in the `internal/static` directory and are embedded with
[`go:embed`](https://pkg.go.dev/embed).

Depending on what your modifications are, it's always a good idea to check that
some of the examples in [./_example](./_example/) work with your modifications
to Statsviz. To do so `cd` to the directory of the example and run:

    go mod edit -replace=github.com/arl/statsviz=../../


## Documentation

No contribution is too small, improvements to code comments and/or README
are welcome!


## Examples

There are many Go libraries to handle HTTP requests, routing, etc..

Feel free to add an example to show how to register Statsviz with your favourite
library.

To do so, please add a directory under `./_example`. For instance, if you want to add an
example showing how to register Statsviz within library `foobar`:

 - create a directory `./_example/foobar/`
 - create a file `./_example/foobar/main.go`
 - call `go example.Work()` as the first line of your example (see other
   examples). This forces the garbage collector to _do something_ so that
   Statsviz interface won't remain static when an user runs your example.
 - the code should be `gofmt`ed
 - the example should compile and run
 - when ran, Statsviz interface should be accessible at http://localhost:8080/debug/statsviz


Thank you!

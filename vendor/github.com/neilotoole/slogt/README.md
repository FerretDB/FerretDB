# slogt

`slogt` is a bridge between Go stdlib [`testing`](https://pkg.go.dev/testing) pkg
and [`log/slog`](https://pkg.go.dev/golang.org/log/slog).


The problem: when tests execute, your `slog` output goes directly to `stdout`,
unlike a call to `t.Log`, which is correlated with your test's execution.

```go
func TestSlog_Ugly(t *testing.T) {
   log := slog.New(slog.NewTextHandler(os.Stdout, nil))
   t.Log("I am indented correctly")
   log.Info("But I am not")
}
```

Produces:

```text
=== RUN   TestSlog_Ugly
    slogt_test.go:22: I am indented correctly
time=2023-04-01T11:29:27.236-06:00 level=INFO msg="But I am not"
```

Note the second line (produced via `slog`).

`slogt` bridges those packages.

```go
func TestSlogt_Pretty(t *testing.T) {
	log := slogt.New(t)
	t.Log("I am indented correctly")
	log.Info("And so am I")
}
```

Produces:

```text
=== RUN   TestSlogt_Pretty
    slogt_test.go:28: I am indented correctly
    logger.go:230: time=2023-04-01T11:33:06.342-06:00 level=INFO msg="And so am I"
```


## Usage

Run `go get` as per procedure:

```shell
go get -u github.com/neilotoole/slogt
```

Then, use `slogt.New` to get a `*slog.Logger` that you can
use as you normally would.

```go
func TestText(t *testing.T) {
   log := slogt.New(t)
   log.Info("hello world")
}
```

Produces:

```text
=== RUN   TestText
    logger.go:230: time=2023-04-01T11:14:53.073-06:00 level=INFO msg="hello world"
```

In practice, you would pass the `*slog.Logger` returned from `slogt.New` to
the component under test. For example:

```go
func TestApp(t *testing.T) {
  log := slogt.New(t)
  
  app := app.New(log, ...) // other dependencies

  result, err := app.DepositMoney(100)
  require.NoError(t, err)
  require.Equal(t, 100, result.Balance)
}
```

If the `app.DepositMoney` method logs anything, its output will be piped
to `t.Log` as desired. 

### Options

The default output is text, i.e. a `slog.TextHandler.` You can
specify JSON using the `slogt.JSON()` option.

```go
func TestJSON(t *testing.T) {
   log := slogt.New(t, slogt.JSON())
   log.Info("hello world")
}
```

Produces:

```text
=== RUN   TestJSON
    logger.go:230: {"time":"2023-04-01T11:14:12.164085-06:00","level":"INFO","msg":"hello world"}
```

To switch the default handler:

```go
func init() {
    slogt.DefaultHandler = slogt.JSON	
}
```

You can exercise full control over the handler using `slogt.Factory()`.

```go
func TestSomething(T *testing.T) {
    // This factory returns a slog.Handler using slog.LevelError.
    f := slogt.Factory(func(w io.Writer) slog.Handler {
       opts := &slog.HandlerOptions{
           Level: slog.LevelError,
       }
       return slog.NewTextHandler(w, opts)
    })

    log := slogt.New(t, f)
}
```

## Deficiency

Calling `t.Log()` prints the callsite as the first element
of each log statement (`logger.go:230` in the example below).

```text
=== RUN   TestText
    logger.go:230: time=2023-04-01T11:14:53.073-06:00 level=INFO msg="hello world"
```

But `logger.go:230` is actually in the internals of `slog` package.
What we really want to see is the location of the caller of, say, `log.Info()`.

Alas, given the available functionality
on `testing.T` (i.e. the `Helper` method), and how `slog` is implemented,
there's no way to have the correct callsite printed.

There are a number of ways this could be fixed:

1. The Go team could implement a `testing.NewLogger(t)` function that effectively
   does what this package does, but it would have access to the `testing.T`'s
   internal state, and so could manipulate the calldepth.
2. The `testing.T` type could expose a `HelperN(depth int)` method that allows
   logging libraries and the like to manipulate the calldepth further. This would
   be generically useful even independent of this particular case.
3. The `slog` package could test if the handler implements an interface with
   method `Helper()`, and if so, invoke that method. This would need to be
   implemented in several spots in `slog` codebase, and would introduce a little
   overhead.

Being that none of the above are available right now, we have to live
with the incorrect callsite always being printed. If you also want to
see the correct callsite alongside the incorrect one, you can do this:

```go
func TestCaller(t *testing.T) {
	f := slogt.Factory(func(w io.Writer) slog.Handler {
		opts := &slog.HandlerOptions{
			AddSource: true,
		}

		return slog.NewTextHandler(w, opts)
	})

	log := slogt.New(t, f)
	log.Info("Show me the real callsite")
}
```

Which produces (note the `source` attribute):

```text
=== RUN   TestCaller
    logger.go:230: time=2023-04-01T11:21:49.896-06:00 level=INFO source=/Users/neilotoole/slogt/slogt_test.go:103 msg="Show me the real callsite"
```



[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=round-square)](https://pkg.go.dev/github.com/arl/statsviz)
[![Latest tag](https://img.shields.io/github/tag/arl/statsviz.svg)](https://github.com/arl/statsviz/tag/)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

[![Test Actions Status](https://github.com/arl/statsviz/workflows/Tests-linux/badge.svg)](https://github.com/arl/statsviz/actions)
[![Test Actions Status](https://github.com/arl/statsviz/workflows/Tests-others/badge.svg)](https://github.com/arl/statsviz/actions)
[![codecov](https://codecov.io/gh/arl/statsviz/branch/main/graph/badge.svg)](https://codecov.io/gh/arl/statsviz)

# Statsviz

<p align="center">
  <img alt="Statsviz Gopher Logo" width="120" src="https://raw.githubusercontent.com/arl/statsviz/readme-docs/logo.png?sanitize=true">
  <img alt="statsviz ui" width="450" align="right" src="https://github.com/arl/statsviz/raw/readme-docs/window.png">
</p>
<br/>

Visualize real time plots of your Go program runtime metrics, including heap, objects, goroutines, GC pauses, scheduler and more, in your browser.

<hr>

- [Statsviz](#statsviz)
  - [Install](#install)
  - [Usage](#usage)
  - [Advanced Usage](#advanced-usage)
  - [How Does That Work?](#how-does-that-work)
  - [Documentation](#documentation)
    - [Go API](#go-api)
    - [User interface](#user-interface)
    - [Plots](#plots)
    - [User Plots](#user-plots)
  - [Examples](#examples)
  - [Questions / Troubleshooting](#questions--troubleshooting)
  - [Contributing](#contributing)
  - [Changelog](#changelog)
  - [License: MIT](#license-mit)

## Install

Download the latest version:

```
go get github.com/arl/statsviz@latest
```

Please note that, as new metrics are added to the `/runtime/metrics` package, new plots are added to Statsviz.
This also means that the presence of some plots on the dashboard depends on the Go version you're using.

When in doubt, use the latest ;-)


## Usage

Register `Statsviz` HTTP handlers with your application `http.ServeMux`.

```go
mux := http.NewServeMux()
statsviz.Register(mux)

go func() {
    log.Println(http.ListenAndServe("localhost:8080", mux))
}()
```

Open your browser at http://localhost:8080/debug/statsviz


## Advanced Usage

If you want more control over Statsviz HTTP handlers, examples are:
 - you're using some HTTP framework
 - you want to place Statsviz handler behind some middleware

then use `statsviz.NewServer` to obtain a `Server` instance. Both the `Index()` and `Ws()` methods return `http.HandlerFunc`.

```go
srv, err := statsviz.NewServer(); // Create server or handle error
srv.Index()                       // UI (dashboard) http.HandlerFunc
srv.Ws()                          // Websocket http.HandlerFunc
```

Please look at examples of usage in the [Examples](_example) directory.


## How Does That Work?

`statsviz.Register` registers 2 HTTP handlers within the given `http.ServeMux`:

- the `Index` handler serves Statsviz user interface at `/debug/statsviz` at the address served by your program.

- The `Ws` serves a Websocket endpoint. When the browser connects to that endpoint, [runtime/metrics](https://pkg.go.dev/runtime/metrics) are sent to the browser, once per second.

Data points are in a browser-side circular-buffer.


## Documentation

### Go API

Check out the API reference on [pkg.go.dev](https://pkg.go.dev/github.com/arl/statsviz#section-documentation).

### User interface

Controls at the top of the page act on all plots:

<img alt="menu" src="https://github.com/arl/statsviz/raw/readme-docs/menu-002.png">

- the groom shows/hides the vertical lines representing garbage collections.
- the time range selector defines the visualized time span.
- the play/pause icons stops and resume the refresh of the plots.
- the light/dark selector switches between light and dark modes.

On top of each plot there are 2 icons:

<img alt="menu" src="https://github.com/arl/statsviz/raw/readme-docs/plot.menu-001.png">

- the camera downloads a PNG image of the plot.
- the info icon shows details about the metrics displayed.

### Plots

Depending on your go version, some plots may not be available.

#### Heap (global)

<img width="50%" alt="heap-global" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/heap-global.png">

#### Heap (details)

<img width="50%" alt="heap-details" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/heap-details.png">

#### Live Objects in Heap

<img width="50%" alt="live-objects" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/live-objects.png">

#### Live Bytes in Heap

<img width="50%" alt="live-bytes" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/live-bytes.png">

#### MSpan/MCache

<img width="50%" alt="mspan-mcache" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/mspan-mcache.png">

#### Memory classes

<img width="50%" alt="memory-classes" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/memory-classes.png">

#### Goroutines

<img width="50%" alt="goroutines" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/goroutines.png">

#### Size Classes

<img width="50%" alt="size-classes" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/size-classes.png">

#### GC Scan

<img width="50%" alt="gc-scan" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/gc-scan.png">

#### GC Cycles

<img width="50%" alt="gc-cycles" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/gc-cycles.png">

#### Stop-the-world Pause Latencies

<img width="50%" alt="gc-pauses" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/gc-pauses.png">

#### CPU Classes (GC)

<img width="50%" alt="cpu-classes-gc" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/cpu-classes-gc.png">

#### Time Goroutines Spend in 'Runnable' state

<img width="50%" alt="runnable-time" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/runnable-time.png">

#### Time Goroutines Spend Blocked on Mutexes

<img width="50%" alt="mutex-wait" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/mutex-wait.png">

#### Starting Size of Goroutines Stacks

<img width="50%" alt="gc-stack-size" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/gc-stack-size.png">

#### Goroutine Scheduling Events

<img width="50%" alt="sched-events" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/sched-events.png">

#### CGO Calls

<img width="50%" alt="cgo" src="https://github.com/arl/statsviz/raw/readme-docs/runtime-metrics/cgo.png">


### User Plots

Since `v0.6` you can add your own plots to Statsviz dashboard, in order to easily
visualize your application metrics next to runtime metrics.

Please see the [userplots example](_example/userplots/main.go).

## Examples

Check out the [\_example](./_example/README.md) directory to see various ways to use Statsviz, such as:

- use of `http.DefaultServeMux` or your own `http.ServeMux`
- wrap HTTP handler behind a middleware
- register the web page at `/foo/bar` instead of `/debug/statsviz`
- use `https://` rather than `http://`
- register Statsviz handlers with various Go HTTP libraries/frameworks:
  - [echo](https://github.com/labstack/echo/)
  - [fasthttp](https://github.com/valyala/fasthttp)
  - [fiber](https://github.com/gofiber/fiber/)
  - [gin](https://github.com/gin-gonic/gin)
  - and many others thanks to many contributors!

## Questions / Troubleshooting

Either use GitHub's [discussions](https://github.com/arl/statsviz/discussions) or come to say hi and ask a live question on [#statsviz channel on Gopher's slack](https://gophers.slack.com/archives/C043DU4NZ9D).

## Contributing

Please use [issues](https://github.com/arl/statsviz/issues/new/choose) for bugs and feature requests.  
Pull-requests are always welcome!  
More details in [CONTRIBUTING.md](CONTRIBUTING.md).

## Changelog

See [CHANGELOG.md](./CHANGELOG.md).

## License: MIT

See [LICENSE](LICENSE)

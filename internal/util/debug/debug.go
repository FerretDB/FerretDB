// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package debug provides debug facilities.
package debug

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	_ "expvar" // for metrics
	"fmt"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof" // for profiling
	"net/url"
	"path/filepath"
	"slices"
	"text/template"
	"time"

	"github.com/arl/statsviz"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

var (
	graphicsPath = "/debug/graphs"
	metricsPath  = "/debug/metrics"
	archivePath  = "/debug/archive"
	pprofPath    = "/debug/pprof"
	varsPath     = "/debug/vars"
)

func generateFileName(prefix, filename string) string {
	return fmt.Sprintf("%s-%s", prefix, filename)
}

// RunHandler runs debug handler.
func RunHandler(ctx context.Context, addr string, r prometheus.Registerer, l *zap.Logger) {
	var err error

	stdL := must.NotFail(zap.NewStdLogAt(l, zap.WarnLevel))

	http.Handle(metricsPath, promhttp.InstrumentMetricHandler(
		r, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			ErrorLog:          stdL,
			ErrorHandling:     promhttp.ContinueOnError,
			Registry:          r,
			EnableOpenMetrics: true,
		}),
	))

	opts := []statsviz.Option{
		statsviz.Root("/debug/graphs"),
		// TODO https://github.com/FerretDB/FerretDB/issues/3600
	}
	must.NoError(statsviz.Register(http.DefaultServeMux, opts...))

	http.HandleFunc(archivePath, func(rw http.ResponseWriter, req *http.Request) {
		var (
			scheme          string
			u               url.URL
			responses       = make(map[string][]byte, 2)
			debugFilePrefix = "FerretDB"
			debugFileName   = "debug.zip"
		)

		u.Path = metricsPath
		u.Host = req.Host

		if req.URL.Scheme == "" {
			scheme = "http"
		}

		u.Scheme = scheme

		metricsFileName := generateFileName(debugFilePrefix, filepath.Base(metricsPath)+".txt")
		responses[metricsFileName], err = performRequest(u)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		u.Path = pprofPath
		pprofFileName := generateFileName(debugFilePrefix, filepath.Base(pprofPath)+".html")
		responses[pprofFileName], err = performRequest(u)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/zip")
		rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", generateFileName(debugFilePrefix, debugFileName)))

		zipWriter := zip.NewWriter(rw)
		defer zipWriter.Close()

		for fileName, response := range responses {
			fileWriter, err := zipWriter.Create(fileName)
			if err != nil {
				l.Error("could not create metrics.txt in zip")
				// Handle error
			}
			_, err = io.Copy(fileWriter, bytes.NewReader(response))
			if err != nil {
				l.Error("could not copy metrics.txt in zip")
				// Handle error
			}
		}
	})

	handlers := map[string]string{
		// custom handlers registered above
		graphicsPath: "Visualize metrics",
		metricsPath:  "Metrics in Prometheus format",
		archivePath:  "Metrics and pprof data in zip format",

		// stdlib handlers
		varsPath:  "Expvar package metrics",
		pprofPath: "Runtime profiling data for pprof",
	}

	var page bytes.Buffer
	must.NoError(template.Must(template.New("debug").Parse(`
	<html>
	<body>
	<ul>
	{{range $path, $desc := .}}
		<li><a href="{{$path}}">{{$path}}</a>: {{$desc}}</li>
	{{end}}
	</ul>
	</body>
	</html>
	`)).Execute(&page, handlers))

	http.HandleFunc("/debug", func(rw http.ResponseWriter, _ *http.Request) {
		rw.Write(page.Bytes())
	})

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		http.Redirect(rw, req, "/debug", http.StatusSeeOther)
	})

	s := http.Server{
		Addr:     addr,
		ErrorLog: stdL,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		lis := must.NotFail(net.Listen("tcp", addr))

		root := fmt.Sprintf("http://%s", lis.Addr())

		l.Sugar().Infof("Starting debug server on %s ...", root)

		paths := maps.Keys(handlers)
		slices.Sort(paths)

		for _, path := range paths {
			l.Sugar().Infof("%s%s - %s", root, path, handlers[path])
		}

		if err := s.Serve(lis); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	<-ctx.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	s.Shutdown(stopCtx) //nolint:contextcheck // use new context for cancellation

	s.Close()
	l.Sugar().Info("Debug server stopped.")
}

func performRequest(u url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Handle response (e.g., read response body)
	defer resp.Body.Close() // Close response body

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}
	return body, nil
}

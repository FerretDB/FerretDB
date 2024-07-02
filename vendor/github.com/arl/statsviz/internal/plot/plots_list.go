package plot

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"runtime/metrics"
	"sync"
	"time"
)

var (
	names       map[string]bool
	usedMetrics map[string]struct{}
)

type plotFunc func(idxs map[string]int) runtimeMetric

var plotFuncs []plotFunc

func registerPlotFunc(f plotFunc) {
	plotFuncs = append(plotFuncs, f)
}

func registerRuntimePlot(name string, metrics ...string) bool {
	if names == nil {
		names = make(map[string]bool)
		usedMetrics = make(map[string]struct{})
	}
	if names[name] {
		panic(name + " is an already reserved plot name")
	}
	names[name] = true

	// Record the metrics we use.
	for _, m := range metrics {
		usedMetrics[m] = struct{}{}
	}

	return true
}

func IsReservedPlotName(name string) bool {
	return names[name]
}

type runtimeMetric interface {
	name() string
	isEnabled() bool
	layout([]metrics.Sample) any
	values([]metrics.Sample) any
}

// List holds all the plots that statsviz knows about. Some plots might be
// disabled, if they rely on metrics that are unknown to the current Go version.
type List struct {
	rtPlots   []runtimeMetric
	userPlots []UserPlot

	once sync.Once // ensure Config is called once
	cfg  *Config

	idxs  map[string]int // map metrics name to idx in samples and descs
	descs []metrics.Description

	mu      sync.Mutex // protects samples in case of concurrent calls to WriteValues
	samples []metrics.Sample
}

func NewList(userPlots []UserPlot) (*List, error) {
	if name := hasDuplicatePlotNames(userPlots); name != "" {
		return nil, fmt.Errorf("duplicate plot name %s", name)
	}

	descs := metrics.All()
	pl := &List{
		idxs:      make(map[string]int),
		descs:     descs,
		samples:   make([]metrics.Sample, len(descs)),
		userPlots: userPlots,
	}
	for i := range pl.samples {
		pl.samples[i].Name = pl.descs[i].Name
		pl.idxs[pl.samples[i].Name] = i
	}
	metrics.Read(pl.samples)

	return pl, nil
}

func (pl *List) Config() *Config {
	pl.once.Do(func() {
		pl.rtPlots = make([]runtimeMetric, 0, len(plotFuncs))
		for _, f := range plotFuncs {
			pl.rtPlots = append(pl.rtPlots, f(pl.idxs))
		}

		layouts := make([]any, 0, len(pl.rtPlots))
		for i := range pl.rtPlots {
			if pl.rtPlots[i].isEnabled() {
				layouts = append(layouts, pl.rtPlots[i].layout(pl.samples))
			}
		}

		pl.cfg = &Config{
			Events: []string{"lastgc"},
			Series: layouts,
		}

		// User plots go at the back of the list for now.
		for i := range pl.userPlots {
			pl.cfg.Series = append(pl.cfg.Series, pl.userPlots[i].Layout())
		}
	})
	return pl.cfg
}

// WriteValues writes into w a JSON object containing the data points for all
// plots at the current instant.
func (pl *List) WriteValues(w io.Writer) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	metrics.Read(pl.samples)

	// lastgc time series is used as source to represent garbage collection
	// timestamps as vertical bars on certain plots.
	gcStats := debug.GCStats{}
	debug.ReadGCStats(&gcStats)

	m := map[string]any{
		// javascript timestampts are in milliseconds
		"lastgc":    []int64{gcStats.LastGC.UnixMilli()},
		"timestamp": time.Now().UnixMilli(),
	}

	for _, p := range pl.rtPlots {
		if p.isEnabled() {
			m[p.name()] = p.values(pl.samples)
		}
	}

	for i := range pl.userPlots {
		up := &pl.userPlots[i]
		switch {
		case up.Scatter != nil:
			vals := make([]float64, len(up.Scatter.Funcs))
			for i := range up.Scatter.Funcs {
				vals[i] = up.Scatter.Funcs[i]()
			}
			m[up.Scatter.Plot.Name] = vals
		case up.Heatmap != nil:
			panic("unimplemented")
		}
	}

	if err := json.NewEncoder(w).Encode(m); err != nil {
		return fmt.Errorf("failed to write/convert metrics values to json: %v", err)
	}
	return nil
}

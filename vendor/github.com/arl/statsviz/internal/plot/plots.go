package plot

import (
	"math"
	"runtime/metrics"
	"time"
)

func init() {
	// lastgc and timestamp are both special cases.

	// lastgc draws vertical lines represeting GCs on certain plots.
	registerRuntimePlot("lastgc")
	// timestamp is the metric is the data for the time axis used on all plots.
	registerRuntimePlot("timestamp")

	registerPlotFunc(makeHeapGlobalPlot)
	registerPlotFunc(makeHeapDetailsPlot)
	registerPlotFunc(makeLiveObjectsPlot)
	registerPlotFunc(makeLiveBytesPlot)
	registerPlotFunc(makeMSpanMCachePlot)
	registerPlotFunc(makeMemoryClassesPlot)
	registerPlotFunc(makeGoroutinesPlot)
	registerPlotFunc(makeSizeClassesPlot)
	registerPlotFunc(makeGCScanPlot)
	registerPlotFunc(makeGCCyclesPlot)
	registerPlotFunc(makeGCPausesPlot)
	registerPlotFunc(makeCPUClassesGCPlot)
	registerPlotFunc(makeRunnableTimePlot)
	registerPlotFunc(makeMutexWaitPlot)
	registerPlotFunc(makeGCStackSizePlot)
	registerPlotFunc(makeSchedEventsPlot)
	registerPlotFunc(makeCGOPlot)
}

/*
 * heap (global)
 */
var _ = registerRuntimePlot("heap-global",
	"/memory/classes/heap/objects:bytes",
	"/memory/classes/heap/unused:bytes",
	"/memory/classes/heap/free:bytes",
	"/memory/classes/heap/released:bytes",
)

type heapGlobal struct {
	enabled bool

	idxobj      int
	idxunused   int
	idxfree     int
	idxreleased int
}

func makeHeapGlobalPlot(idxs map[string]int) runtimeMetric {
	idxobj, ok1 := idxs["/memory/classes/heap/objects:bytes"]
	idxunused, ok2 := idxs["/memory/classes/heap/unused:bytes"]
	idxfree, ok3 := idxs["/memory/classes/heap/free:bytes"]
	idxreleased, ok4 := idxs["/memory/classes/heap/released:bytes"]

	return &heapGlobal{
		enabled:     ok1 && ok2 && ok3 && ok4,
		idxobj:      idxobj,
		idxunused:   idxunused,
		idxfree:     idxfree,
		idxreleased: idxreleased,
	}
}

func (p *heapGlobal) name() string    { return "heap-global" }
func (p *heapGlobal) isEnabled() bool { return p.enabled }

func (p *heapGlobal) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:   p.name(),
		Title:  "Heap (global)",
		Type:   "scatter",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:       "heap in-use",
				Unitfmt:    "%{y:.4s}B",
				HoverOn:    "points+fills",
				StackGroup: "one",
			},
			{
				Name:       "heap free",
				Unitfmt:    "%{y:.4s}B",
				HoverOn:    "points+fills",
				StackGroup: "one",
			},
			{
				Name:       "heap released",
				Unitfmt:    "%{y:.4s}B",
				HoverOn:    "points+fills",
				StackGroup: "one",
			},
		},
		InfoText: `<i>Heap in use</i> is <b>/memory/classes/heap/objects + /memory/classes/heap/unused</b>. It amounts to the memory occupied by live objects and dead objects that are not yet marked free by the GC, plus some memory reserved for heap objects.
<i>Heap free</i> is <b>/memory/classes/heap/free</b>, that is free memory that could be returned to the OS, but has not been.
<i>Heap released</i> is <b>/memory/classes/heap/free</b>, memory that is free memory that has been returned to the OS.`,
	}
	s.Layout.Yaxis.TickSuffix = "B"
	s.Layout.Yaxis.Title = "bytes"
	return s
}

func (p *heapGlobal) values(samples []metrics.Sample) any {
	heapObjects := samples[p.idxobj].Value.Uint64()
	heapUnused := samples[p.idxunused].Value.Uint64()

	heapInUse := heapObjects + heapUnused
	heapFree := samples[p.idxfree].Value.Uint64()
	heapReleased := samples[p.idxreleased].Value.Uint64()
	return []uint64{
		heapInUse,
		heapFree,
		heapReleased,
	}
}

/*
 * heap (details)
 */
var _ = registerRuntimePlot("heap-details",
	"/memory/classes/heap/objects:bytes",
	"/memory/classes/heap/unused:bytes",
	"/memory/classes/heap/free:bytes",
	"/memory/classes/heap/released:bytes",
	"/memory/classes/heap/stacks:bytes",
	"/gc/heap/goal:bytes",
)

type heapDetails struct {
	enabled bool

	idxobj      int
	idxunused   int
	idxfree     int
	idxreleased int
	idxstacks   int
	idxgoal     int
}

func makeHeapDetailsPlot(idxs map[string]int) runtimeMetric {
	idxobj, ok1 := idxs["/memory/classes/heap/objects:bytes"]
	idxunused, ok2 := idxs["/memory/classes/heap/unused:bytes"]
	idxfree, ok3 := idxs["/memory/classes/heap/free:bytes"]
	idxreleased, ok4 := idxs["/memory/classes/heap/released:bytes"]
	idxstacks, ok5 := idxs["/memory/classes/heap/stacks:bytes"]
	idxgoal, ok6 := idxs["/gc/heap/goal:bytes"]

	return &heapDetails{
		enabled:     ok1 && ok2 && ok3 && ok4 && ok5 && ok6,
		idxobj:      idxobj,
		idxunused:   idxunused,
		idxfree:     idxfree,
		idxreleased: idxreleased,
		idxstacks:   idxstacks,
		idxgoal:     idxgoal,
	}
}

func (p *heapDetails) name() string    { return "heap-details" }
func (p *heapDetails) isEnabled() bool { return p.enabled }

func (p *heapDetails) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:   p.name(),
		Title:  "Heap (details)",
		Type:   "scatter",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:    "heap sys",
				Unitfmt: "%{y:.4s}B",
			},
			{
				Name:    "heap objects",
				Unitfmt: "%{y:.4s}B",
			},
			{
				Name:    "heap stacks",
				Unitfmt: "%{y:.4s}B",
			},
			{
				Name:    "heap goal",
				Unitfmt: "%{y:.4s}B",
			},
		},
		InfoText: `<i>Heap</i> sys is <b>/memory/classes/heap/objects + /memory/classes/heap/unused + /memory/classes/heap/released + /memory/classes/heap/free</b>. It's an estimate of all the heap memory obtained form the OS.
<i>Heap objects</i> is <b>/memory/classes/heap/objects</b>, the memory occupied by live objects and dead objects that have not yet been marked free by the GC.
<i>Heap stacks</i> is <b>/memory/classes/heap/stacks</b>, the memory used for stack space.
<i>Heap goal</i> is <b>gc/heap/goal</b>, the heap size target for the end of the GC cycle.`,
	}
	s.Layout.Yaxis.TickSuffix = "B"
	s.Layout.Yaxis.Title = "bytes"
	return s
}

func (p *heapDetails) values(samples []metrics.Sample) any {
	heapObjects := samples[p.idxobj].Value.Uint64()
	heapUnused := samples[p.idxunused].Value.Uint64()

	heapInUse := heapObjects + heapUnused
	heapFree := samples[p.idxfree].Value.Uint64()
	heapReleased := samples[p.idxreleased].Value.Uint64()

	heapIdle := heapReleased + heapFree
	heapSys := heapInUse + heapIdle
	heapStacks := samples[p.idxstacks].Value.Uint64()
	nextGC := samples[p.idxgoal].Value.Uint64()

	return []uint64{
		heapSys,
		heapObjects,
		heapStacks,
		nextGC,
	}
}

/*
 * live objects
 */
var _ = registerRuntimePlot("live-objects", "/gc/heap/objects:objects")

type liveObjects struct {
	enabled bool

	idxobjects int
}

func makeLiveObjectsPlot(idxs map[string]int) runtimeMetric {
	idxobjects, ok := idxs["/gc/heap/objects:objects"]

	return &liveObjects{
		enabled:    ok,
		idxobjects: idxobjects,
	}
}

func (p *liveObjects) name() string    { return "live-objects" }
func (p *liveObjects) isEnabled() bool { return p.enabled }

func (p *liveObjects) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:   p.name(),
		Title:  "Live Objects in Heap",
		Type:   "bar",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:    "live objects",
				Unitfmt: "%{y:.4s}",
				Color:   RGBString(255, 195, 128),
			},
		},
		InfoText: `<i>Live objects</i> is <b>/gc/heap/objects</b>. It's the number of objects, live or unswept, occupying heap memory.`,
	}
	s.Layout.Yaxis.Title = "objects"
	return s
}

func (p *liveObjects) values(samples []metrics.Sample) any {
	gcHeapObjects := samples[p.idxobjects].Value.Uint64()
	return []uint64{
		gcHeapObjects,
	}
}

/*
 * live bytes
 */
var _ = registerRuntimePlot("live-bytes",
	"/gc/heap/allocs:bytes",
	"/gc/heap/frees:bytes",
)

type liveBytes struct {
	enabled bool

	idxallocs int
	idxfrees  int
}

func makeLiveBytesPlot(idxs map[string]int) runtimeMetric {
	idxallocs, ok1 := idxs["/gc/heap/allocs:bytes"]
	idxfrees, ok2 := idxs["/gc/heap/frees:bytes"]

	return &liveBytes{
		enabled:   ok1 && ok2,
		idxallocs: idxallocs,
		idxfrees:  idxfrees,
	}
}

func (p *liveBytes) name() string    { return "live-bytes" }
func (p *liveBytes) isEnabled() bool { return p.enabled }

func (p *liveBytes) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:   p.name(),
		Title:  "Live Bytes in Heap",
		Type:   "bar",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:    "live bytes",
				Unitfmt: "%{y:.4s}B",
				Color:   RGBString(135, 182, 218),
			},
		},
		InfoText: `<i>Live bytes</i> is <b>/gc/heap/allocs - /gc/heap/frees</b>. It's the number of bytes currently allocated (and not yet GC'ec) to the heap by the application.`,
	}
	s.Layout.Yaxis.Title = "bytes"
	return s
}

func (p *liveBytes) values(samples []metrics.Sample) any {
	allocBytes := samples[p.idxallocs].Value.Uint64()
	freedBytes := samples[p.idxfrees].Value.Uint64()
	return []uint64{
		allocBytes - freedBytes,
	}
}

/*
 * mspan mcache
 */
var _ = registerRuntimePlot("mspan-mcache",
	"/memory/classes/metadata/mspan/inuse:bytes",
	"/memory/classes/metadata/mspan/free:bytes",
	"/memory/classes/metadata/mcache/inuse:bytes",
	"/memory/classes/metadata/mcache/free:bytes",
)

type mspanMcache struct {
	enabled bool

	idxmspanInuse  int
	idxmspanFree   int
	idxmcacheInuse int
	idxmcacheFree  int
}

func makeMSpanMCachePlot(idxs map[string]int) runtimeMetric {
	idxmspanInuse, ok1 := idxs["/memory/classes/metadata/mspan/inuse:bytes"]
	idxmspanFree, ok2 := idxs["/memory/classes/metadata/mspan/free:bytes"]
	idxmcacheInuse, ok3 := idxs["/memory/classes/metadata/mcache/inuse:bytes"]
	idxmcacheFree, ok4 := idxs["/memory/classes/metadata/mcache/free:bytes"]

	return &mspanMcache{
		enabled:        ok1 && ok2 && ok3 && ok4,
		idxmspanInuse:  idxmspanInuse,
		idxmspanFree:   idxmspanFree,
		idxmcacheInuse: idxmcacheInuse,
		idxmcacheFree:  idxmcacheFree,
	}
}

func (p *mspanMcache) name() string    { return "mspan-mcache" }
func (p *mspanMcache) isEnabled() bool { return p.enabled }

func (p *mspanMcache) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:   p.name(),
		Title:  "MSpan/MCache",
		Type:   "scatter",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:    "mspan in-use",
				Unitfmt: "%{y:.4s}B",
			},
			{
				Name:    "mspan free",
				Unitfmt: "%{y:.4s}B",
			},
			{
				Name:    "mcache in-use",
				Unitfmt: "%{y:.4s}B",
			},
			{
				Name:    "mcache free",
				Unitfmt: "%{y:.4s}B",
			},
		},
		InfoText: `<i>Mspan in-use</i> is <b>/memory/classes/metadata/mspan/inuse</b>, the memory that is occupied by runtime mspan structures that are currently being used.
<i>Mspan free</i> is <b>/memory/classes/metadata/mspan/free</b>, the memory that is reserved for runtime mspan structures, but not in-use.
<i>Mcache in-use</i> is <b>/memory/classes/metadata/mcache/inuse</b>, the memory that is occupied by runtime mcache structures that are currently being used.
<i>Mcache free</i> is <b>/memory/classes/metadata/mcache/free</b>, the memory that is reserved for runtime mcache structures, but not in-use.
`,
	}
	s.Layout.Yaxis.Title = "bytes"
	s.Layout.Yaxis.TickSuffix = "B"
	return s
}

func (p *mspanMcache) values(samples []metrics.Sample) any {
	mspanInUse := samples[p.idxmspanInuse].Value.Uint64()
	mspanSys := samples[p.idxmspanFree].Value.Uint64()
	mcacheInUse := samples[p.idxmcacheInuse].Value.Uint64()
	mcacheSys := samples[p.idxmcacheFree].Value.Uint64()
	return []uint64{
		mspanInUse,
		mspanSys,
		mcacheInUse,
		mcacheSys,
	}
}

/*
 * goroutines
 */
var _ = registerRuntimePlot("goroutines", "/sched/goroutines:goroutines")

type goroutines struct {
	enabled bool

	idxgs int
}

func makeGoroutinesPlot(idxs map[string]int) runtimeMetric {
	idxgs, ok := idxs["/sched/goroutines:goroutines"]

	return &goroutines{
		enabled: ok,
		idxgs:   idxgs,
	}
}

func (p *goroutines) name() string    { return "goroutines" }
func (p *goroutines) isEnabled() bool { return p.enabled }

func (p *goroutines) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:   p.name(),
		Title:  "Goroutines",
		Type:   "scatter",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:    "goroutines",
				Unitfmt: "%{y}",
			},
		},
		InfoText: "<i>Goroutines</i> is <b>/sched/goroutines</b>, the count of live goroutines.",
	}

	s.Layout.Yaxis.Title = "goroutines"
	return s
}

func (p *goroutines) values(samples []metrics.Sample) any {
	return []uint64{samples[p.idxgs].Value.Uint64()}
}

/*
 * size classes
 */
var _ = registerRuntimePlot("size-classes",
	"/gc/heap/allocs-by-size:bytes",
	"/gc/heap/frees-by-size:bytes",
)

type sizeClasses struct {
	enabled     bool
	sizeClasses []uint64

	idxallocs int
	idxfrees  int
}

func makeSizeClassesPlot(idxs map[string]int) runtimeMetric {
	idxallocs, ok1 := idxs["/gc/heap/allocs-by-size:bytes"]
	idxfrees, ok2 := idxs["/gc/heap/frees-by-size:bytes"]

	return &sizeClasses{
		enabled:   ok1 && ok2,
		idxallocs: idxallocs,
		idxfrees:  idxfrees,
	}
}

func (p *sizeClasses) name() string    { return "size-classes" }
func (p *sizeClasses) isEnabled() bool { return p.enabled }

func (p *sizeClasses) layout(samples []metrics.Sample) any {
	// Perform a sanity check on the number of buckets on the 'allocs' and
	// 'frees' size classes histograms. Statsviz plots a single histogram based
	// on those 2 so we want them to have the same number of buckets, which
	// should be true.
	allocsBySize := samples[p.idxallocs].Value.Float64Histogram()
	freesBySize := samples[p.idxfrees].Value.Float64Histogram()
	if len(allocsBySize.Buckets) != len(freesBySize.Buckets) {
		panic("different number of buckets in allocs and frees size classes histograms!")
	}

	// Pre-allocate here so we never do it in values.
	p.sizeClasses = make([]uint64, len(allocsBySize.Counts))

	// No downsampling for the size classes histogram (factor=1) but we still
	// need to adapt boundaries for plotly heatmaps.
	buckets := downsampleBuckets(allocsBySize, 1)

	h := Heatmap{
		Name:       p.name(),
		Title:      "Size Classes",
		Type:       "heatmap",
		UpdateFreq: 5,
		Colorscale: BlueShades,
		Buckets:    floatseq(len(buckets)),
		CustomData: buckets,
		Hover: HeapmapHover{
			YName: "size class",
			YUnit: "bytes",
			ZName: "objects",
		},
		InfoText: `This heatmap shows the distribution of size classes, using <b>/gc/heap/allocs-by-size</b> and <b>/gc/heap/frees-by-size</b>.`,
		Layout: HeatmapLayout{
			YAxis: HeatmapYaxis{
				Title:    "size class",
				TickMode: "array",
				TickVals: []float64{1, 9, 17, 25, 31, 37, 43, 50, 58, 66},
				TickText: []float64{1 << 4, 1 << 7, 1 << 8, 1 << 9, 1 << 10, 1 << 11, 1 << 12, 1 << 13, 1 << 14, 1 << 15},
			},
		},
	}
	return h
}

func (p *sizeClasses) values(samples []metrics.Sample) any {
	allocsBySize := samples[p.idxallocs].Value.Float64Histogram()
	freesBySize := samples[p.idxfrees].Value.Float64Histogram()

	for i := 0; i < len(p.sizeClasses); i++ {
		p.sizeClasses[i] = allocsBySize.Counts[i] - freesBySize.Counts[i]
	}
	return p.sizeClasses
}

/*
 * gc pauses
 */
var _ = registerRuntimePlot("gc-pauses", "/gc/pauses:seconds")

type gcpauses struct {
	enabled    bool
	histfactor int
	counts     [maxBuckets]uint64

	idxgcpauses int
}

func makeGCPausesPlot(idxs map[string]int) runtimeMetric {
	idxgcpauses, ok := idxs["/gc/pauses:seconds"]

	return &gcpauses{
		enabled:     ok,
		idxgcpauses: idxgcpauses,
	}
}

func (p *gcpauses) name() string    { return "gc-pauses" }
func (p *gcpauses) isEnabled() bool { return p.enabled }

func (p *gcpauses) layout(samples []metrics.Sample) any {
	gcpauses := samples[p.idxgcpauses].Value.Float64Histogram()
	p.histfactor = downsampleFactor(len(gcpauses.Buckets), maxBuckets)
	buckets := downsampleBuckets(gcpauses, p.histfactor)

	h := Heatmap{
		Name:       p.name(),
		Title:      "Stop-the-world Pause Latencies",
		Type:       "heatmap",
		UpdateFreq: 5,
		Colorscale: PinkShades,
		Buckets:    floatseq(len(buckets)),
		CustomData: buckets,
		Hover: HeapmapHover{
			YName: "pause duration",
			YUnit: "duration",
			ZName: "pauses",
		},
		Layout: HeatmapLayout{
			YAxis: HeatmapYaxis{
				Title:    "pause duration",
				TickMode: "array",
				TickVals: []float64{6, 13, 20, 26, 33, 39.5, 46, 53, 60, 66, 73, 79, 86},
				TickText: []float64{1e-7, 1e-6, 1e-5, 1e-4, 1e-3, 5e-3, 1e-2, 5e-2, 1e-1, 5e-1, 1, 5, 10},
			},
		},
		InfoText: `This heatmap shows the distribution of individual GC-related stop-the-world pause latencies, uses <b>/gc/pauses:seconds</b>,.`,
	}
	return h
}

func (p *gcpauses) values(samples []metrics.Sample) any {
	gcpauses := samples[p.idxgcpauses].Value.Float64Histogram()
	return downsampleCounts(gcpauses, p.histfactor, p.counts[:])
}

/*
 * time spent in runnable state
 */
var _ = registerRuntimePlot("runnable-time", "/sched/latencies:seconds")

type runnableTime struct {
	enabled    bool
	histfactor int
	counts     [maxBuckets]uint64

	idxschedlat int
}

func makeRunnableTimePlot(idxs map[string]int) runtimeMetric {
	idxschedlat, ok := idxs["/sched/latencies:seconds"]

	return &runnableTime{
		enabled:     ok,
		idxschedlat: idxschedlat,
	}
}

func (p *runnableTime) name() string    { return "runnable-time" }
func (p *runnableTime) isEnabled() bool { return p.enabled }

func (p *runnableTime) layout(samples []metrics.Sample) any {
	schedlat := samples[p.idxschedlat].Value.Float64Histogram()
	p.histfactor = downsampleFactor(len(schedlat.Buckets), maxBuckets)
	buckets := downsampleBuckets(schedlat, p.histfactor)

	h := Heatmap{
		Name:       p.name(),
		Title:      "Time Goroutines Spend in 'Runnable' state",
		Type:       "heatmap",
		UpdateFreq: 5,
		Colorscale: GreenShades,
		Buckets:    floatseq(len(buckets)),
		CustomData: buckets,
		Hover: HeapmapHover{
			YName: "duration",
			YUnit: "duration",
			ZName: "goroutines",
		},
		Layout: HeatmapLayout{
			YAxis: HeatmapYaxis{
				Title:    "duration",
				TickMode: "array",
				TickVals: []float64{6, 13, 20, 26, 33, 39.5, 46, 53, 60, 66, 73, 79, 86},
				TickText: []float64{1e-7, 1e-6, 1e-5, 1e-4, 1e-3, 5e-3, 1e-2, 5e-2, 1e-1, 5e-1, 1, 5, 10},
			},
		},
		InfoText: `This heatmap shows the distribution of the time goroutines have spent in the scheduler in a runnable state before actually running, uses <b>/sched/latencies:seconds</b>.`,
	}

	return h
}

func (p *runnableTime) values(samples []metrics.Sample) any {
	schedlat := samples[p.idxschedlat].Value.Float64Histogram()

	return downsampleCounts(schedlat, p.histfactor, p.counts[:])
}

/*
 * scheduling events
 */
var _ = registerRuntimePlot("sched-events",
	"/sched/latencies:seconds",
	"/sched/gomaxprocs:threads",
)

type schedEvents struct {
	enabled bool

	idxschedlat   int
	idxGomaxprocs int
	lasttot       uint64
}

func makeSchedEventsPlot(idxs map[string]int) runtimeMetric {
	idxschedlat, ok1 := idxs["/sched/latencies:seconds"]
	idxGomaxprocs, ok2 := idxs["/sched/gomaxprocs:threads"]

	return &schedEvents{
		enabled:       ok1 && ok2,
		idxschedlat:   idxschedlat,
		idxGomaxprocs: idxGomaxprocs,
		lasttot:       math.MaxUint64,
	}
}

func (p *schedEvents) name() string    { return "sched-events" }
func (p *schedEvents) isEnabled() bool { return p.enabled }

func (p *schedEvents) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:   p.name(),
		Title:  "Goroutine Scheduling Events",
		Type:   "scatter",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:    "events per unit of time",
				Unitfmt: "%{y}",
			},
			{
				Name:    "events per unit of time, per P",
				Unitfmt: "%{y}",
			},
		},
		InfoText: `<i>Events per second</i> is the sum of all buckets in <b>/sched/latencies:seconds</b>, that is, it tracks the total number of goroutine scheduling events. That number is multiplied by the constant 8.
<i>Events per second per P (processor)</i> is <i>Events per second</i> divided by current <b>GOMAXPROCS</b>, from <b>/sched/gomaxprocs:threads</b>.
<b>NOTE</b>: the multiplying factor comes from internal Go runtime source code and might change from version to version.`,
	}
	s.Layout.Yaxis.Title = "events"
	return s
}

// gTrackingPeriod is currently always 8. Guard it behind build tags when that
// changes. See https://github.com/golang/go/blob/go1.18.4/src/runtime/runtime2.go#L502-L504
const currentGtrackingPeriod = 8

// TODO show scheduling events per seconds
func (p *schedEvents) values(samples []metrics.Sample) any {
	schedlat := samples[p.idxschedlat].Value.Float64Histogram()
	gomaxprocs := samples[p.idxGomaxprocs].Value.Uint64()

	total := uint64(0)
	for _, v := range schedlat.Counts {
		total += v
	}
	total *= currentGtrackingPeriod

	curtot := total - p.lasttot
	if p.lasttot == math.MaxUint64 {
		// We don't want a big spike at statsviz launch in case the process has
		// been running for some time and curtot is high.
		curtot = 0
	}
	p.lasttot = total

	ftot := float64(curtot)

	return []float64{
		ftot,
		ftot / float64(gomaxprocs),
	}
}

/*
 * cgo
 */
var _ = registerRuntimePlot("cgo", "/cgo/go-to-c-calls:calls")

type cgo struct {
	enabled  bool
	idxgo2c  int
	lastgo2c uint64
}

func makeCGOPlot(idxs map[string]int) runtimeMetric {
	idxgo2c, ok := idxs["/cgo/go-to-c-calls:calls"]

	return &cgo{
		enabled:  ok,
		idxgo2c:  idxgo2c,
		lastgo2c: math.MaxUint64,
	}
}

func (p *cgo) name() string    { return "cgo" }
func (p *cgo) isEnabled() bool { return p.enabled }

func (p *cgo) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:  p.name(),
		Title: "CGO Calls",
		Type:  "bar",
		Subplots: []Subplot{
			{
				Name:    "calls from go to c",
				Unitfmt: "%{y}",
				Color:   "red",
			},
		},
		InfoText: "Shows the count of calls made from Go to C by the current process, per unit of time. Uses <b>/cgo/go-to-c-calls:calls</b>",
	}

	s.Layout.Yaxis.Title = "calls"
	return s
}

// TODO show cgo calls per second
func (p *cgo) values(samples []metrics.Sample) any {
	go2c := samples[p.idxgo2c].Value.Uint64()
	curgo2c := go2c - p.lastgo2c
	if p.lastgo2c == math.MaxUint64 {
		curgo2c = 0
	}
	p.lastgo2c = go2c

	return []uint64{curgo2c}
}

/*
 * gc stack size
 */
var _ = registerRuntimePlot("gc-stack-size", "/gc/stack/starting-size:bytes")

type gcStackSize struct {
	enabled  bool
	idxstack int
}

func makeGCStackSizePlot(idxs map[string]int) runtimeMetric {
	idxstack, ok := idxs["/gc/stack/starting-size:bytes"]

	return &gcStackSize{
		enabled:  ok,
		idxstack: idxstack,
	}
}

func (p *gcStackSize) name() string    { return "gc-stack-size" }
func (p *gcStackSize) isEnabled() bool { return p.enabled }

func (p *gcStackSize) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:  p.name(),
		Title: "Starting Size of Goroutines Stacks",
		Type:  "scatter",
		Subplots: []Subplot{
			{
				Name:    "new goroutines stack size",
				Unitfmt: "%{y:.4s}B",
			},
		},
		InfoText: "Shows the stack size of new goroutines, uses <b>/gc/stack/starting-size:bytes</b>",
	}

	s.Layout.Yaxis.Title = "bytes"
	return s
}

func (p *gcStackSize) values(samples []metrics.Sample) any {
	stackSize := samples[p.idxstack].Value.Uint64()
	return []uint64{stackSize}
}

/*
 * GC cycles
 */
var _ = registerRuntimePlot("gc-cycles",
	"/gc/cycles/automatic:gc-cycles",
	"/gc/cycles/forced:gc-cycles",
	"/gc/cycles/total:gc-cycles",
)

type gcCycles struct {
	enabled bool

	idxAutomatic int
	idxForced    int
	idxTotal     int

	lastAuto, lastForced, lastTotal uint64
}

func makeGCCyclesPlot(idxs map[string]int) runtimeMetric {
	idxAutomatic, ok1 := idxs["/gc/cycles/automatic:gc-cycles"]
	idxForced, ok2 := idxs["/gc/cycles/forced:gc-cycles"]
	idxTotal, ok3 := idxs["/gc/cycles/total:gc-cycles"]

	return &gcCycles{
		enabled:      ok1 && ok2 && ok3,
		idxAutomatic: idxAutomatic,
		idxForced:    idxForced,
		idxTotal:     idxTotal,
	}
}

func (p *gcCycles) name() string    { return "gc-cycles" }
func (p *gcCycles) isEnabled() bool { return p.enabled }

func (p *gcCycles) layout(_ []metrics.Sample) any {
	return Scatter{
		Name:  p.name(),
		Title: "Completed GC Cycles",
		Type:  "bar",
		Subplots: []Subplot{
			{
				Name:    "automatic",
				Unitfmt: "%{y}",
				Type:    "bar",
			},
			{
				Name:    "forced",
				Unitfmt: "%{y}",
				Type:    "bar",
			},
		},
		InfoText: `Number of completed GC cycles, either forced of generated by the Go runtime.`,
		Layout: ScatterLayout{
			BarMode: "stack",
			Yaxis: ScatterYAxis{
				Title: "cycles",
			},
		},
	}
}

func (p *gcCycles) values(samples []metrics.Sample) any {
	total := samples[p.idxTotal].Value.Uint64()
	auto := samples[p.idxAutomatic].Value.Uint64()
	forced := samples[p.idxForced].Value.Uint64()

	if p.lastTotal == 0 {
		p.lastTotal = total
		p.lastForced = forced
		p.lastAuto = auto
		return []uint64{0, 0}
	}

	ret := []uint64{
		auto - p.lastAuto,
		forced - p.lastForced,
	}

	p.lastForced = forced
	p.lastAuto = auto

	return ret
}

/*
* mspan mcache
 */
var _ = registerRuntimePlot("memory-classes",
	"/memory/classes/os-stacks:bytes",
	"/memory/classes/other:bytes",
	"/memory/classes/profiling/buckets:bytes",
	"/memory/classes/total:bytes",
)

type memoryClasses struct {
	enabled bool

	idxOSStacks    int
	idxOther       int
	idxProfBuckets int
	idxTotal       int
}

func makeMemoryClassesPlot(idxs map[string]int) runtimeMetric {
	idxOSStacks, ok1 := idxs["/memory/classes/os-stacks:bytes"]
	idxOther, ok2 := idxs["/memory/classes/other:bytes"]
	idxProfBuckets, ok3 := idxs["/memory/classes/profiling/buckets:bytes"]
	idxTotal, ok4 := idxs["/memory/classes/total:bytes"]

	return &memoryClasses{
		enabled:        ok1 && ok2 && ok3 && ok4,
		idxOSStacks:    idxOSStacks,
		idxOther:       idxOther,
		idxProfBuckets: idxProfBuckets,
		idxTotal:       idxTotal,
	}
}

func (p *memoryClasses) name() string    { return "memory-classes" }
func (p *memoryClasses) isEnabled() bool { return p.enabled }

func (p *memoryClasses) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:   p.name(),
		Title:  "Memory classes",
		Type:   "scatter",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:    "os stacks",
				Unitfmt: "%{y:.4s}B",
			},
			{
				Name:    "other",
				Unitfmt: "%{y:.4s}B",
			},
			{
				Name:    "profiling buckets",
				Unitfmt: "%{y:.4s}B",
			},
			{
				Name:    "total",
				Unitfmt: "%{y:.4s}B",
			},
		},

		InfoText: `<i>OS stacks</i> is <b>/memory/classes/os-stacks</b>, stack memory allocated by the underlying operating system.
<i>Other</i> is <b>/memory/classes/other</b>, memory used by execution trace buffers, structures for debugging the runtime, finalizer and profiler specials, and more.
<i>Profiling buckets</i> is <b>/memory/classes/profiling/buckets</b>, memory that is used by the stack trace hash map used for profiling.
<i>Total</i> is <b>/memory/classes/total</b>, all memory mapped by the Go runtime into the current process as read-write.`,
	}
	s.Layout.Yaxis.Title = "bytes"
	s.Layout.Yaxis.TickSuffix = "B"
	return s
}

func (p *memoryClasses) values(samples []metrics.Sample) any {
	osStacks := samples[p.idxOSStacks].Value.Uint64()
	other := samples[p.idxOther].Value.Uint64()
	profBuckets := samples[p.idxProfBuckets].Value.Uint64()
	total := samples[p.idxTotal].Value.Uint64()

	return []uint64{
		osStacks,
		other,
		profBuckets,
		total,
	}
}

/*
* cpu classes (gc)
 */
var _ = registerRuntimePlot("cpu-classes-gc",
	"/cpu/classes/gc/mark/assist:cpu-seconds",
	"/cpu/classes/gc/mark/dedicated:cpu-seconds",
	"/cpu/classes/gc/mark/idle:cpu-seconds",
	"/cpu/classes/gc/pause:cpu-seconds",
	"/cpu/classes/gc/total:cpu-seconds",
)

type cpuClassesGC struct {
	enabled bool

	idxMarkAssist    int
	idxMarkDedicated int
	idxMarkIdle      int
	idxPause         int
	idxTotal         int

	lastTime time.Time

	lastMarkAssist    float64
	lastMarkDedicated float64
	lastMarkIdle      float64
	lastPause         float64
	lastTotal         float64
}

func makeCPUClassesGCPlot(idxs map[string]int) runtimeMetric {
	idxMarkAssist, ok1 := idxs["/cpu/classes/gc/mark/assist:cpu-seconds"]
	idxMarkDedicated, ok2 := idxs["/cpu/classes/gc/mark/dedicated:cpu-seconds"]
	idxMarkIdle, ok3 := idxs["/cpu/classes/gc/mark/idle:cpu-seconds"]
	idxPause, ok4 := idxs["/cpu/classes/gc/pause:cpu-seconds"]
	idxTotal, ok5 := idxs["/cpu/classes/gc/total:cpu-seconds"]

	return &cpuClassesGC{
		enabled:          ok1 && ok2 && ok3 && ok4 && ok5,
		idxMarkAssist:    idxMarkAssist,
		idxMarkDedicated: idxMarkDedicated,
		idxMarkIdle:      idxMarkIdle,
		idxPause:         idxPause,
		idxTotal:         idxTotal,
	}
}

func (p *cpuClassesGC) name() string    { return "cpu-classes-gc" }
func (p *cpuClassesGC) isEnabled() bool { return p.enabled }

func (p *cpuClassesGC) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:   p.name(),
		Title:  "CPU classes (GC)",
		Type:   "scatter",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:    "mark assist",
				Unitfmt: "%{y:.4s}s",
			},
			{
				Name:    "mark dedicated",
				Unitfmt: "%{y:.4s}s",
			},
			{
				Name:    "mark idle",
				Unitfmt: "%{y:.4s}s",
			},
			{
				Name:    "pause",
				Unitfmt: "%{y:.4s}s",
			},
			{
				Name:    "total",
				Unitfmt: "%{y:.4s}s",
			},
		},

		InfoText: `Cumulative metrics are converted to rates by Statsviz so as to be more easily comparable and readable.
All this metrics are overestimates, and not directly comparable to system CPU time measurements. Compare only with other /cpu/classes metrics.

<i>mark assist</i> is <b>/cpu/classes/gc/mark/assist</b>, estimated total CPU time goroutines spent performing GC tasks to assist the GC and prevent it from falling behind the application.
<i>mark dedicated</i> is <b>/cpu/classes/gc/mark/dedicated</b>, Estimated total CPU time spent performing GC tasks on processors (as defined by GOMAXPROCS) dedicated to those tasks.
<i>mark idle</i> is <b>/cpu/classes/gc/mark/idle</b>, estimated total CPU time spent performing GC tasks on spare CPU resources that the Go scheduler could not otherwise find a use for.
<i>pause</i> is <b>/cpu/classes/gc/pause</b>, estimated total CPU time spent with the application paused by the GC.
<i>total</i> is <b>/cpu/classes/gc/total</b>, estimated total CPU time spent performing GC tasks.`,
	}
	s.Layout.Yaxis.Title = "cpu-seconds per seconds"
	s.Layout.Yaxis.TickSuffix = "s"
	return s
}

func (p *cpuClassesGC) values(samples []metrics.Sample) any {
	curMarkAssist := samples[p.idxMarkAssist].Value.Float64()
	curMarkDedicated := samples[p.idxMarkDedicated].Value.Float64()
	curMarkIdle := samples[p.idxMarkIdle].Value.Float64()
	curPause := samples[p.idxPause].Value.Float64()
	curTotal := samples[p.idxTotal].Value.Float64()

	if p.lastTime.IsZero() {
		p.lastMarkAssist = curMarkAssist
		p.lastMarkDedicated = curMarkDedicated
		p.lastMarkIdle = curMarkIdle
		p.lastPause = curPause
		p.lastTotal = curTotal
		p.lastTime = time.Now()

		return []float64{0, 0, 0, 0, 0}
	}

	t := time.Since(p.lastTime).Seconds()

	markAssist := (curMarkAssist - p.lastMarkAssist) / t
	markDedicated := (curMarkDedicated - p.lastMarkDedicated) / t
	markIdle := (curMarkIdle - p.lastMarkIdle) / t
	pause := (curPause - p.lastPause) / t
	total := (curTotal - p.lastTotal) / t

	p.lastMarkAssist = curMarkAssist
	p.lastMarkDedicated = curMarkDedicated
	p.lastMarkIdle = curMarkIdle
	p.lastPause = curPause
	p.lastTotal = curTotal
	p.lastTime = time.Now()

	return []float64{
		markAssist,
		markDedicated,
		markIdle,
		pause,
		total,
	}
}

/*
* mutex wait
 */
var _ = registerRuntimePlot("mutex-wait",
	"/sync/mutex/wait/total:seconds",
)

type mutexWait struct {
	enabled      bool
	idxMutexWait int

	lastTime      time.Time
	lastMutexWait float64
}

func makeMutexWaitPlot(idxs map[string]int) runtimeMetric {
	idxMutexWait, ok := idxs["/cpu/classes/gc/mark/assist:cpu-seconds"]

	return &mutexWait{
		enabled:      ok,
		idxMutexWait: idxMutexWait,
	}
}

func (p *mutexWait) name() string    { return "mutex-wait" }
func (p *mutexWait) isEnabled() bool { return p.enabled }

func (p *mutexWait) layout(_ []metrics.Sample) any {
	s := Scatter{
		Name:   p.name(),
		Title:  "Time Goroutines Spend Blocked on Mutexes",
		Type:   "scatter",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:    "mutex wait",
				Unitfmt: "%{y:.4s}s",
			},
		},

		InfoText: `Cumulative metrics are converted to rates by Statsviz so as to be more easily comparable and readable.
<i>mutex wait</i> is <b>/sync/mutex/wait/total</b>, approximate cumulative time goroutines have spent blocked on a sync.Mutex or sync.RWMutex.

This metric is useful for identifying global changes in lock contention. Collect a mutex or block profile using the runtime/pprof package for more detailed contention data.`,
	}
	s.Layout.Yaxis.Title = "seconds per seconds"
	s.Layout.Yaxis.TickSuffix = "s"
	return s
}

func (p *mutexWait) values(samples []metrics.Sample) any {
	if p.lastTime.IsZero() {
		p.lastTime = time.Now()
		p.lastMutexWait = samples[p.idxMutexWait].Value.Float64()

		return []float64{0}
	}

	t := time.Since(p.lastTime).Seconds()

	mutexWait := (samples[p.idxMutexWait].Value.Float64() - p.lastMutexWait) / t

	p.lastMutexWait = samples[p.idxMutexWait].Value.Float64()
	p.lastTime = time.Now()

	return []float64{
		mutexWait,
	}
}

/*
 * gc scan
 */
var _ = registerRuntimePlot("gc-scan",
	"/gc/scan/globals:bytes",
	"/gc/scan/heap:bytes",
	"/gc/scan/stack:bytes",
	"/gc/scan/total:bytes",
)

type gcScan struct {
	enabled bool

	idxGlobals int
	idxHeap    int
	idxStack   int
}

func makeGCScanPlot(idxs map[string]int) runtimeMetric {
	idxGlobals, ok1 := idxs["/gc/scan/globals:bytes"]
	idxHeap, ok2 := idxs["/gc/scan/heap:bytes"]
	idxStack, ok3 := idxs["/gc/scan/stack:bytes"]

	return &gcScan{
		enabled:    ok1 && ok2 && ok3,
		idxGlobals: idxGlobals,
		idxHeap:    idxHeap,
		idxStack:   idxStack,
	}
}

func (p *gcScan) name() string    { return "gc-scan" }
func (p *gcScan) isEnabled() bool { return p.enabled }

func (p *gcScan) layout(_ []metrics.Sample) any {
	return Scatter{
		Name:   p.name(),
		Title:  "GC Scan",
		Type:   "bar",
		Events: "lastgc",
		Subplots: []Subplot{
			{
				Name:    "scannable globals",
				Unitfmt: "%{y:.4s}B",
				Type:    "bar",
			},
			{
				Name:    "scannable heap",
				Unitfmt: "%{y:.4s}B",
				Type:    "bar",
			},
			{
				Name:    "scanned stack",
				Unitfmt: "%{y:.4s}B",
				Type:    "bar",
			},
		},
		InfoText: `This plot shows the amount of memory that is scannable by the GC.
<i>scannable globals</i> is <b>/gc/scan/globals</b>, the total amount of global variable space that is scannable.
<i>scannable heap</i> is <b>/gc/scan/heap</b>, the total amount of heap space that is scannable.
<i>scanned stack</i> is <b>/gc/scan/stack</b>, the number of bytes of stack that were scanned last GC cycle.
`,
		Layout: ScatterLayout{
			BarMode: "stack",
			Yaxis: ScatterYAxis{
				TickSuffix: "B",
				Title:      "bytes",
			},
		},
	}
}

func (p *gcScan) values(samples []metrics.Sample) any {
	globals := samples[p.idxGlobals].Value.Uint64()
	heap := samples[p.idxHeap].Value.Uint64()
	stack := samples[p.idxStack].Value.Uint64()
	return []uint64{
		globals,
		heap,
		stack,
	}
}

/*
 * helpers
 */

func floatseq(n int) []float64 {
	seq := make([]float64, n)
	for i := 0; i < n; i++ {
		seq[i] = float64(i)
	}
	return seq
}

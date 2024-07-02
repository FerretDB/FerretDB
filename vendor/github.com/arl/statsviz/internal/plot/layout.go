package plot

type (
	Config struct {
		// Series contains the plots we want to show and how we want to show them.
		Series []any `json:"series"`
		// Events contains a list of 'events time series' names. Series with
		// these names must be sent alongside other series. An event time series
		// is just made of timestamps with no associated value, each of which
		// gets plotted as a vertical line over another plot.
		Events []string `json:"events"`
	}

	Scatter struct {
		Name       string        `json:"name"`
		Title      string        `json:"title"`
		Type       string        `json:"type"`
		UpdateFreq int           `json:"updateFreq"`
		InfoText   string        `json:"infoText"`
		Events     string        `json:"events"`
		Layout     ScatterLayout `json:"layout"`
		Subplots   []Subplot     `json:"subplots"`
	}
	ScatterLayout struct {
		Yaxis   ScatterYAxis `json:"yaxis"`
		BarMode string       `json:"barmode"`
	}
	ScatterYAxis struct {
		Title      string `json:"title"`
		TickSuffix string `json:"ticksuffix"`
	}

	Subplot struct {
		Name       string `json:"name"`
		Unitfmt    string `json:"unitfmt"`
		StackGroup string `json:"stackgroup"`
		HoverOn    string `json:"hoveron"`
		Color      string `json:"color"`
		Type       string `json:"type"`
	}

	Heatmap struct {
		Name       string          `json:"name"`
		Title      string          `json:"title"`
		Type       string          `json:"type"`
		UpdateFreq int             `json:"updateFreq"`
		InfoText   string          `json:"infoText"`
		Events     string          `json:"events"`
		Layout     HeatmapLayout   `json:"layout"`
		Colorscale []WeightedColor `json:"colorscale"`
		Buckets    []float64       `json:"buckets"`
		CustomData []float64       `json:"custom_data"`
		Hover      HeapmapHover    `json:"hover"`
	}
	HeatmapLayout struct {
		YAxis HeatmapYaxis `json:"yaxis"`
	}
	HeatmapYaxis struct {
		Title    string    `json:"title"`
		TickMode string    `json:"tickmode"`
		TickVals []float64 `json:"tickvals"`
		TickText []float64 `json:"ticktext"`
	}
	HeapmapHover struct {
		YName string `json:"yname"`
		YUnit string `json:"yunit"` // 'duration', 'bytes' or custom
		ZName string `json:"zname"`
	}
)

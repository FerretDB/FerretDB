package plot

type ScatterUserPlot struct {
	Plot  Scatter
	Funcs []func() float64
}

type HeatmapUserPlot struct {
	Plot Heatmap
	// TODO(arl): heatmap get value func
}

type UserPlot struct {
	Scatter *ScatterUserPlot
	Heatmap *HeatmapUserPlot
}

func (up UserPlot) Layout() any {
	switch {
	case (up.Scatter != nil) == (up.Heatmap != nil):
		panic("userplot must be a timeseries or a heatmap")
	case up.Scatter != nil:
		return up.Scatter.Plot
	case up.Heatmap != nil:
		return up.Heatmap.Plot
	}

	panic("unreachable")
}

func hasDuplicatePlotNames(userPlots []UserPlot) string {
	names := map[string]bool{}
	for _, p := range userPlots {
		name := ""
		if p.Scatter != nil {
			name = p.Scatter.Plot.Name
		} else if p.Heatmap != nil {
			name = p.Heatmap.Plot.Name
		} else {
			panic("both heapmap and scatter are nil")
		}
		if names[name] {
			return name
		}
		names[name] = true
	}
	return ""
}

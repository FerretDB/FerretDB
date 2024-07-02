package statsviz

import (
	"errors"
	"fmt"

	"github.com/arl/statsviz/internal/plot"
)

// TimeSeriesType describes the type of a time series plot.
type TimeSeriesType string

const (
	// Scatter is a time series plot made of lines.
	Scatter TimeSeriesType = "scatter"

	// Bar is a time series plot made of bars.
	Bar TimeSeriesType = "bar"
)

// BarMode determines how bars at the same location are displayed on a bar plot.
type BarMode string

const (
	// Stack indicates that bars are stacked on top of one another.
	Stack BarMode = "stack"

	// Ggroup indicates that bars are plotted next to one another, centered
	// around the shared location.
	Group BarMode = "group"

	// Relative indicates that bars are stacked on top of one another, with
	// negative values below the axis and positive values above.
	Relative BarMode = "relative"

	// Overlay indicates that bars are plotted over one another.
	Overlay BarMode = "overlay"
)

var (
	// ErrNoTimeSeries is returned when a user plot has no time series.
	ErrNoTimeSeries = errors.New("user plot must have at least one time series")

	// ErrEmptyPlotName is returned when a user plot has an empty name.
	ErrEmptyPlotName = errors.New("user plot name can't be empty")
)

// ErrReservedPlotName is returned when a reserved plot name is used for a user plot.
type ErrReservedPlotName string

func (e ErrReservedPlotName) Error() string {
	return fmt.Sprintf("%q is a reserved plot name", string(e))
}

// HoverOnType describes the type of hover effect on a time series plot.
type HoverOnType string

const (
	// HoverOnPoints specifies that the hover effects highlights individual
	// points.
	HoverOnPoints HoverOnType = "points"

	// HoverOnPoints specifies that the hover effects highlights filled regions.
	HoverOnFills HoverOnType = "fills"

	// HoverOnPointsAndFills specifies that the hover effects highlights both
	// points and filled regions.
	HoverOnPointsAndFills HoverOnType = "points+fills"
)

// A TimeSeries describes a single time series of a plot.
type TimeSeries struct {
	// Name is the name identifying this time series in the user interface.
	Name string

	// UnitFmt is the d3-format string used to format the numbers of this time
	// series in the user interface. See https://github.com/d3/d3-format.
	Unitfmt string

	// HoverOn configures whether the hover effect highlights individual points
	// or do they highlight filled regions, or both. Defaults to HoverOnFills.
	HoverOn HoverOnType

	// Type is the time series type, either [Scatter] or [Bar]. default: [Scatter].
	Type TimeSeriesType

	// GetValue specifies the function called to get the value of this time
	// series.
	GetValue func() float64
}

// TimeSeriesPlotConfig describes the configuration of a time series plot.
type TimeSeriesPlotConfig struct {
	// Name is the plot name, it must be unique.
	Name string

	// Title is the plot title, shown above the plot.
	Title string

	// Type is either [Scatter] or [Bar]. default: [Scatter].
	Type TimeSeriesType

	// BarMode is either [Stack], [Group], [Relative] or [Overlay].
	// default: [Group].
	BarMode BarMode

	// Tooltip is the html-aware text shown when the user clicks on the plot
	// Info icon.
	InfoText string

	// YAxisTitle is the title of Y axis.
	YAxisTitle string

	// YAxisTickSuffix is the suffix added to tick values.
	YAxisTickSuffix string

	// Series contains the time series shown on this plot, there must be at
	// least one.
	Series []TimeSeries
}

// Build validates the configuration and builds a time series plot for it
func (p TimeSeriesPlotConfig) Build() (TimeSeriesPlot, error) {
	var zero TimeSeriesPlot
	if p.Name == "" {
		return zero, ErrEmptyPlotName
	}
	if plot.IsReservedPlotName(p.Name) {
		return zero, ErrReservedPlotName(p.Name)
	}
	if len(p.Series) == 0 {
		return zero, ErrNoTimeSeries
	}

	var (
		subplots []plot.Subplot
		funcs    []func() float64
	)
	for _, ts := range p.Series {
		switch ts.HoverOn {
		case "":
			ts.HoverOn = HoverOnFills
		case HoverOnPoints, HoverOnFills, HoverOnPointsAndFills:
			// ok
		default:
			return zero, fmt.Errorf("invalid HoverOn value %s", ts.HoverOn)
		}

		subplots = append(subplots, plot.Subplot{
			Name:    ts.Name,
			Unitfmt: ts.Unitfmt,
			HoverOn: string(ts.HoverOn),
			Type:    string(ts.Type),
		})
		funcs = append(funcs, ts.GetValue)
	}

	return TimeSeriesPlot{
		timeseries: &plot.ScatterUserPlot{
			Plot: plot.Scatter{
				Name:     p.Name,
				Title:    p.Title,
				Type:     string(p.Type),
				InfoText: p.InfoText,
				Layout: plot.ScatterLayout{
					BarMode: string(p.BarMode),
					Yaxis: plot.ScatterYAxis{
						Title:      p.YAxisTitle,
						TickSuffix: p.YAxisTickSuffix,
					},
				},
				Subplots: subplots,
			},
			Funcs: funcs,
		},
	}, nil
}

// TimeSeriesPlot is an opaque type representing a timeseries plot.
// A plot can be created with [TimeSeriesPlotConfig.Build].
type TimeSeriesPlot struct {
	timeseries *plot.ScatterUserPlot
}

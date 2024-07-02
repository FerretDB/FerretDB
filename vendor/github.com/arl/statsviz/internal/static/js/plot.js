import * as theme from "./theme.js";

var infoIcon = {
    'width': 470,
    'height': 530,
    'path': 'M256 0C114.6 0 0 114.6 0 256s114.6 256 256 256s256-114.6 256-256S397.4 0 256 0zM256 128c17.67 0 32 14.33 32 32c0 17.67-14.33 32-32 32S224 177.7 224 160C224 142.3 238.3 128 256 128zM296 384h-80C202.8 384 192 373.3 192 360s10.75-24 24-24h16v-64H224c-13.25 0-24-10.75-24-24S210.8 224 224 224h32c13.25 0 24 10.75 24 24v88h16c13.25 0 24 10.75 24 24S309.3 384 296 384z'
}

const newConfigObject = (cfg) => {
    return {
        // showEditInChartStudio: true,
        // plotlyServerURL: "https://chart-studio.plotly.com",
        displaylogo: false,
        modeBarButtonsToRemove: ['2D', 'zoom2d', 'pan2d', 'select2d', 'lasso2d', 'zoomIn2d', 'zoomOut2d', 'autoScale2d', 'resetScale2d', 'toggleSpikelines'],
        modeBarButtonsToAdd: [{
            name: 'info',
            title: "Plot info",
            icon: infoIcon,
            val: false,
            click: handleInfoButton,
        },],
        toImageButtonOptions: {
            format: 'png',
            filename: cfg.name,
        }
    }
}

const copyArrayOrNull = (o) => {
    return Array.isArray(o) && [...o] || null;
}

const newLayoutObject = (cfg) => {
    const layout = {
        title: {
            y: 0.88,
            font: {
                family: "Roboto",
                size: 18,
            },
            text: cfg.title,
        },
        margin: {
            t: 80,
        },
        paper_bgcolor: cfg.layout.paper_bgcolor,
        plot_bgcolor: cfg.layout.plot_bgcolor,
        font: {
            color: cfg.layout.font_color
        },
        width: 630,
        height: 450,
        hovermode: 'x',
        barmode: cfg.layout.barmode,
        xaxis: {
            tickformat: '%H:%M:%S',
            type: "date",
            fixedrange: true,
            autorange: false,
        },
        yaxis: {
            exponentformat: 'SI',
            tickmode: cfg.layout.yaxis.tickmode,
            ticktext: copyArrayOrNull(cfg.layout.yaxis.ticktext),
            tickvals: copyArrayOrNull(cfg.layout.yaxis.tickvals),
            title: cfg.layout.yaxis.title,
            ticksuffix: cfg.layout.yaxis.ticksuffix,
            fixedrange: true,
        },
        showlegend: true,
        legend: {
            "orientation": "h"
        }
    };

    if (layout.yaxis.tickmode == "array") {
        // Format yaxis ticks
        const formatYUnit = formatFunction(cfg.hover.yunit);
        for (let i = 0; i < layout.yaxis.ticktext.length; i++) {
            layout.yaxis.ticktext[i] = formatYUnit(layout.yaxis.ticktext[i]);
        }
    }

    return layout;
}

const handleInfoButton = (gd, ev) => {
    let button = ev.currentTarget;
    let val = (button.getAttribute('data-val') === 'true');

    const options = {
        allowHTML: true,
        trigger: 'click',
    };

    const instance = tippy(ev.currentTarget, options);
    instance.setContent("<div>" + gd.infoText + "</div>");
    if (val) {
        instance.hide();
    } else {
        instance.show();
    }
    button.setAttribute('data-val', !val);
}

const themeColors = {
    light: {
        paper_bgcolor: '#f8f8f8',
        plot_bgcolor: '#ffffdd',
        font_color: '#434343'
    },
    dark: {
        paper_bgcolor: '#181a1c',
        plot_bgcolor: '#282a2c',
        font_color: '#fff'
    }
};

/*
    Plot configuration object:
    {
      "name": string,                  // internal name
      "title": string,                 // plot title 
      "type": 'scatter'|'bar'|'heatmap' 
      "updateFreq": int,               // datapoints to receive before redrawing the plot. (default: 1)
      "infoText": string,              // text showed in the plot 'info' tooltip
      "events": "lastgc",              // source of vertical lines (example: 'lastgc')
      "layout": object,                // (depends on plot type)
      "subplots": array,               // describe 'traces', only for 'scatter' or 'bar' plots
      "heatmap": object,               // heatmap details
     }

    Layout for 'scatter' and 'bar' plots:
    {
        "yaxis": {
            "title": {
                "text": "bytes"      // yaxis title
            },
            "ticksuffix": "B",       // base unit for ticks
        },
        "barmode": "stack",           // 'stack' or 'group' (only for bar plots)
    },

    Layout" for heatmaps:
    {
        "yaxis": {
            tickmode:  string  (supports 'array' only)
            tickvals:  []float64
            ticktext:  []float64
            "title": {
                "text": "size class"
            }
    }

    Subplots show the potentially multiple trace objects for 'scatter' and 'bar'
    plots. Each trace is an object:
    {
        "name": string;          // internal name
        "unitfmt": string,       // d3 format string for tooltip
        "stackgroup": string,    // stackgroup (if stacked line any)
        "hoveron": string        // useful for stacked only (TODO(arl): remove from go)
        "color": colorstring     // plot/trace color
    }

    Heatmap details object
    {
         "colorscale": array      // array of weighted colors,
         "buckets": array
         "hover": {
             "yname": string,     // y axis units
             "yunit": "bytes",    // y axis name
             "zname": "objects"   // z axis name 
         }
     }
*/


class Plot {
    /**
     * Construct a new Plot object, wrapping a Plotly chart. See above
     * documentation for plot configuration.
     */

    constructor(cfg) {
        cfg.layout.paper_bgcolor = themeColors[theme.getThemeMode()].paper_bgcolor;
        cfg.layout.plot_bgcolor = themeColors[theme.getThemeMode()].plot_bgcolor;
        cfg.layout.font_color = themeColors[theme.getThemeMode()].font_color;

        this._cfg = cfg;
        this._updateCount = 0;
        this._dataTemplate = [];
        this._lastData = [{ x: new Date() }];

        if (this._cfg.type == 'heatmap') {
            this._dataTemplate.push({
                type: 'heatmap',
                x: null,
                y: this._cfg.buckets,
                z: null,
                showlegend: false,
                colorscale: this._cfg.colorscale,
                custom_data: this._cfg.custom_data,
            });
        } else {
            this._cfg.subplots.forEach(subplot => {
                this._dataTemplate.push({
                    type: this._cfg.type,
                    x: null,
                    y: null,
                    name: subplot.name,
                    hovertemplate: `<b>${subplot.unitfmt}</b>`,
                })
            });
        }

        this._plotlyLayout = newLayoutObject(cfg);
        this._plotlyConfig = newConfigObject(cfg);
    }

    name() {
        return this._cfg.name;
    }

    createElement(div, idx) {
        this._htmlElt = div;
        this._plotIdx = idx;
        // Pass a single data with no data to create an empty plot, this removes
        // the 'bad time formatting' warning at startup.
        Plotly.newPlot(this._htmlElt, this._lastData, this._plotlyLayout, this._plotlyConfig);
        if (this._cfg.type == 'heatmap') {
            this._installHeatmapTooltip();
        }

        this._htmlElt.infoText = this._cfg.infoText.split('\n').map(line => `<p>${line}</p>`).join('');
    }

    _installHeatmapTooltip() {
        const options = {
            followCursor: true,
            trigger: "manual",
            allowHTML: true
        };
        const instance = tippy(document.body, options);
        const hover = this._cfg.hover;
        const formatYUnit = formatFunction(hover.yunit);

        const onHover = (data) => {
            const pt2txt = (d) => {
                let bucket;
                if (d.y == 0) {
                    const yhigh = formatYUnit(d.data.custom_data[d.y]);
                    bucket = `(-Inf, ${yhigh})`;
                } else if (d.y == d.data.custom_data.length - 1) {
                    const ylow = formatYUnit(d.data.custom_data[d.y]);
                    bucket = `[${ylow}, +Inf)`;
                } else {
                    const ylow = formatYUnit(d.data.custom_data[d.y - 1]);
                    const yhigh = formatYUnit(d.data.custom_data[d.y]);
                    bucket = `[${ylow}, ${yhigh})`;
                }

                return `
<div class="tooltip-table tooltip-style">
    <div class="tooltip-row">
        <div class="tooltip-label">${hover.yname}</div>
        <div class="tooltip-value">${bucket}</div>
    </div>
    <div class="tooltip-row">
        <div class="tooltip-label">${hover.zname}</div>
        <div class="tooltip-value">${d.z}</div>
    </div>
</div> `;
            }
            instance.setContent(data.points.map(pt2txt)[0]);
            instance.show();
        };
        const onUnhover = (data) => {
            instance.hide();
        };

        this._htmlElt.on('plotly_hover', onHover)
            .on('plotly_unhover', onUnhover);
    }

    _extractData(data) {
        const serie = data.series.get(this._cfg.name);
        if (this._cfg.type == 'heatmap') {
            this._dataTemplate[0].x = data.times;
            this._dataTemplate[0].z = serie;
            this._dataTemplate[0].hoverinfo = 'none';
        } else {
            for (let i = 0; i < this._dataTemplate.length; i++) {
                this._dataTemplate[i].x = data.times;
                this._dataTemplate[i].y = serie[i];
                this._dataTemplate[i].stackgroup = this._cfg.subplots[i].stackgroup;
                this._dataTemplate[i].hoveron = this._cfg.subplots[i].hoveron;
                this._dataTemplate[i].type = this._cfg.subplots[i].type || this._cfg.type;
                this._dataTemplate[i].marker = {
                    color: this._cfg.subplots[i].color,
                };
            }
        }
        return this._dataTemplate;
    }

    update(xrange, data, shapes) {
        this._lastData = this._extractData(data);
        this._updateCount++;
        if (this._cfg.updateFreq == 0 || (this._updateCount % this._cfg.updateFreq == 0)) {
            // Update layout with vertical shapes if necessary.
            if (this._cfg.events != '') {
                this._plotlyLayout.shapes = shapes.get(this._cfg.events);
            }

            // Move the xaxis time range.
            this._plotlyLayout.xaxis.range = xrange;

            Plotly.react(this._htmlElt, this._lastData, this._plotlyLayout, this._plotlyConfig);
        }
    }

    /**
     * update theme color and immediately force plot redraw to apply the new theme
     */
    updateTheme() {
        const themeMode = theme.getThemeMode();
        this._cfg.layout.paper_bgcolor = themeColors[themeMode].paper_bgcolor;
        this._cfg.layout.plot_bgcolor = themeColors[themeMode].plot_bgcolor;
        this._cfg.layout.font_color = themeColors[themeMode].font_color;

        this._plotlyLayout = newLayoutObject(this._cfg);
        this._plotlyConfig = newConfigObject(this._cfg);

        Plotly.react(this._htmlElt, this._lastData, this._plotlyLayout);
    }
};

// Create 'vertical lines' shapes for each of the given timestamps.
const createVerticalLines = (tss) => {
    const shapes = [];
    for (let i = 0, n = tss.length; i < n; i++) {
        const d = tss[i];
        shapes.push({
            type: 'line',
            x0: d,
            x1: d,
            yref: 'paper',
            y0: 0,
            y1: 1,
            line: {
                color: 'rgb(55, 128, 191)',
                width: 1,
                dash: 'longdashdot',
            }
        })
    }
    return shapes;
}

export { createVerticalLines, Plot };

const durUnits = ['w', 'd', 'h', 'm', 's', 'ms', 'µs', 'ns'];
const durVals = [6048e11, 864e11, 36e11, 6e10, 1e9, 1e6, 1e3, 1];

// Formats a time duration provided in second.
const formatDuration = sec => {
    let ns = sec * 1e9;
    for (let i = 0; i < durUnits.length; i++) {
        let inc = ns / durVals[i];

        if (inc < 1) continue;
        return Math.round(inc) + durUnits[i];
    }
    return res.trim();
};

const bytesUnits = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB'];

// Formats a size in bytes.
const formatBytes = bytes => {
    let i = 0;
    while (bytes > 1000) {
        bytes /= 1000;
        i++;
    }
    const res = Math.trunc(bytes);
    return `${res}${bytesUnits[i]}`;
};

// Returns a format function based on the provided unit.
const formatFunction = unit => {
    switch (unit) {
        case 'duration':
            return formatDuration;
        case 'bytes':
            return formatBytes;
    }
    // Default formatting
    return (y) => { `${y} ${hover.yunit}` };
};

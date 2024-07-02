import * as stats from './stats.js';
import * as plot from "./plot.js";
import * as theme from "./theme.js";
import PlotsDef from './plotsdef.js';

const buildWebsocketURI = () => {
    var loc = window.location,
        ws_prot = "ws:";
    if (loc.protocol === "https:") {
        ws_prot = "wss:";
    }
    return ws_prot + "//" + loc.host + loc.pathname + "ws"
}

const dataRetentionSeconds = 600;
var timeout = 250;

const clamp = (val, min, max) => {
    if (val < min) return min;
    if (val > max) return max;
    return val;
}

/* nav bar ui management */
let paused = false;
let show_gc = true;
let timerange = 60;

/* WebSocket connection handling */
const connect = () => {
    const uri = buildWebsocketURI();
    let ws = new WebSocket(uri);
    console.info(`Attempting websocket connection to server at ${uri}`);

    ws.onopen = () => {
        console.info("Successfully connected");
        timeout = 250; // reset connection timeout for next time
    };

    ws.onclose = event => {
        console.error(`Closed websocket connection: code ${event.code}`);
        setTimeout(connect, clamp(timeout += timeout, 250, 5000));
    };

    ws.onerror = err => {
        console.error(`Websocket error, closing connection.`);
        ws.close();
    };

    let initDone = false;
    ws.onmessage = event => {
        let data = JSON.parse(event.data)

        if (!initDone) {
            configurePlots(PlotsDef);
            stats.init(PlotsDef, dataRetentionSeconds);

            attachPlots();

            $('#play_pause').change(() => { paused = !paused; });
            $('#show_gc').change(() => {
                show_gc = !show_gc;
                updatePlots();
            });
            $('#select_timerange').click(() => {
                const val = parseInt($("#select_timerange option:selected").val(), 10);
                timerange = val;
                updatePlots();
            });
            initDone = true;
            return;
        }

        stats.pushData(data);
        if (paused) {
            return
        }
        updatePlots(PlotsDef.events);
    }
}

connect();

let plots = [];

const configurePlots = (plotdefs) => {
    plots = [];
    plotdefs.series.forEach(plotdef => {
        plots.push(new plot.Plot(plotdef));
    });
}

const attachPlots = () => {
    let plotsDiv = $('#plots');
    plotsDiv.empty();

    for (let i = 0; i < plots.length; i++) {
        const plot = plots[i];
        let div = $(`<div id="${plot.name()}">`);
        plot.createElement(div[0], i)
        plotsDiv.append(div);
    }
}

const updatePlots = () => {
    // Create shapes.
    let shapes = new Map();

    let data = stats.slice(timerange);

    if (show_gc) {
        for (const [name, serie] of data.events) {
            shapes.set(name, plot.createVerticalLines(serie));
        }
    }

    // Always show the full range on x axis.
    const now = data.times[data.times.length - 1];
    let xrange = [now - timerange * 1000, now];

    plots.forEach(plot => {
        if (!plot.hidden) {
            plot.update(xrange, data, shapes);
        }
    });
}

const updatePlotsLayout = () => {
    plots.forEach(plot => {
        plot.updateTheme();
    });
}

theme.updateThemeMode();

/**
 * Change color theme when the user presses the theme switch button
 */
$('#color_theme_sw').change(() => {
    const themeMode = theme.getThemeMode();
    const newTheme = themeMode === "dark" && "light" || "dark";
    localStorage.setItem("theme-mode", newTheme);    
    theme.updateThemeMode();
    updatePlotsLayout();
});

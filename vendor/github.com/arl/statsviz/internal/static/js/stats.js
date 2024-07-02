// stats holds the data and function to modify it.
import Buffer from "./buffer.js";

var series = {
    times: null,
    eventsData: new Map(),
    plotData: new Map(),
};

// initialize time series storage.
const init = (plotdefs, buflen) => {
    const extraBufferCapacity = 20; // 20% of extra (preallocated) buffer datapoints
    const bufcap = buflen + (buflen * extraBufferCapacity) / 100; // number of actual datapoints

    series.times = new Buffer(buflen, bufcap);
    series.plotData.clear();
    plotdefs.series.forEach(plotdef => {
        let ndim;
        switch (plotdef.type) {
            case 'bar':
            case 'scatter':
                ndim = plotdef.subplots.length;
                break;
            case 'heatmap':
                ndim = plotdef.buckets.length;
                break;
            default:
                console.error(`[statsviz]: unknown plot type "${plotdef.type}"`);
                return;
        };

        let data = new Array(ndim);
        for (let i = 0; i < ndim; i++) {
            data[i] = new Buffer(buflen, bufcap);
        }
        series.plotData.set(plotdef.name, data);
    });

    plotdefs.events.forEach(event => {
        series.eventsData.set(event, new Array());
    });
}

// push a new datapoint to all time series.
const pushData = (data) => {
    series.times.push(data.timestamp);

    // Update time series.
    for (const [name, plotData] of series.plotData) {
        const curdata = data[name];
        for (let i = 0; i < curdata.length; i++) {
            plotData[i].push(curdata[i]);
        }
    }

    // Update events series, deduplicating event timestamps and trimming the ones
    // that are oldest with respect to the oldest timestamp we're keeping track of.
    for (const [name, event] of series.eventsData) {
        if (event.length == 0) {
            if (data[name].length != 0) {
                const eventTs = new Date(Math.floor(data[name][0]));
                event.push(eventTs);
            }
            return;
        }
        const eventTs = new Date(Math.floor(data[name][0]));
        if (eventTs.getTime() != event[event.length - 1].getTime()) {
            event.push(eventTs);
            let mints = series.times._buf[0];
            if (event[0] < mints) {
                event.splice(0, 1);
            }
        }
    }
}

// slice returns the last n items from all time series.
const slice = (n) => {
    let sliced = {
        times: series.times.slice(n),
        series: new Map(),
        events: series.eventsData,
    };

    for (const [name, plotData] of series.plotData) {
        const arr = new Array(plotData.length);
        for (let i = 0; i < plotData.length; i++) {
            arr[i] = plotData[i].slice(n);
        }
        sliced.series.set(name, arr);
    }
    return sliced;
}

export { init, pushData, slice };
package plot

import (
	"math"
	"runtime/metrics"
)

// maxBuckets is the maximum number of buckets we'll plots in heatmaps.
// Histograms with more buckets than that are going to be downsampled.
const maxBuckets = 100

// downsampleFactor computes the downsampling factor to use with
// downsampleCounts and downsampleBuckets. nstart and nfinal are the number of
// buckets in the original and final bucket.
func downsampleFactor(norg, nfinal int) int {
	mod := norg % nfinal
	if mod == 0 {
		return norg / nfinal
	}
	return 1 + norg/nfinal
}

// downsampleBuckets downsamples the number of buckets in the provided
// histogram, using the given dividing factor, and returns a slice of bucket
// widths.
//
// Given that metrics.Float64Histogram contains the boundaries of histogram
// buckets, the first bucket is not even considered since we're only interested
// in upper bounds. Also, since we can't draw an infinitely large bucket, if h
// last bucket holds +Inf, the width of the last returned bucket will
// extrapolated from the previous 2 buckets.
func downsampleBuckets(h *metrics.Float64Histogram, factor int) []float64 {
	var ret []float64
	vals := h.Buckets[1:]

	for i := 0; i < len(vals); i++ {
		if (i+1)%factor == 0 {
			ret = append(ret, vals[i])
		}
	}
	if len(vals)%factor != 0 {
		// If the number of bucket is not divisible by the factor, let's make a
		// last downsampled bucket, even if it doesn't 'contain' the same number
		// of original buckets.
		ret = append(ret, vals[len(vals)-1])
	}

	if len(ret) > 2 && math.IsInf(ret[len(ret)-1], 1) {
		// Plotly doesn't support a +Inf bound for the last bucket. So we make it
		// so that the last bucket has the same 'width' than the penultimate one.
		ret[len(ret)-1] = ret[len(ret)-2] - ret[len(ret)-3] + ret[len(ret)-2]
	}

	return ret
}

// downsampleCounts downsamples the counts in the provided histogram, using the
// given factor. Every 'factor' buckets are merged into one, larger, bucket. If
// the number of buckets is not divisible by 'factor', then an additional last
// bucket will contain the sum of the counts in all relainbing buckets.
//
// Note: slice should be a slice of maxBuckets elements, so that it can be
// reused across calls.
func downsampleCounts(h *metrics.Float64Histogram, factor int, slice []uint64) []uint64 {
	vals := h.Counts

	if factor == 1 {
		copy(slice, vals)
		slice = slice[:len(vals)]
		return slice
	}

	slice = slice[:0]

	var sum uint64
	for i := 0; i < len(vals); i++ {
		if i%factor == 0 && i > 1 {
			slice = append(slice, sum)
			sum = vals[i]
		} else {
			sum += vals[i]
		}
	}

	// Whatever sum remains, it goes to the last bucket.
	return append(slice, sum)
}

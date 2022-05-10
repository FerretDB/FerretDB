package logging

import "container/ring"

var logram = ring.New(1024)

func getLogRam() *ring.Ring {
	return logram
}

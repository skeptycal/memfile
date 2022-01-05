package memfile

import (
	"sync"
	"time"
)

var tracerStartTime = time.Now()

type smap = sync.Map

type (
	// snapshot is an individual snapshot triggered by Snap().
	// snapshot contain configurable data related to the
	// instant the snapshot was recorded.
	snapshot struct{}

	// timeLine is an ordered map of snapshots
	timeLine map[time.Duration]snapshot

	Tracer struct {
		Name      string    // Name of the Tracer run
		StartTime time.Time // the zero time for snapshot deltas
		timeline  timeLine  // a timeline including snapshots
	}
)

func init() {

}

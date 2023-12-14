package main

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/pinger"
	"sync"
	"time"
)

var (
	stlock          sync.Mutex
	status_trackers = make(map[uint64]*status)
)

type status struct {
	ID    uint64
	state bool
	since time.Time
}

func get_status_tracker(ID uint64) *status {
	stlock.Lock()
	defer stlock.Unlock()
	res, found := status_trackers[ID]
	if found {
		return res
	}
	res = &status{ID: ID, since: time.Now()}
	status_trackers[ID] = res
	return res
}
func (s *status) Set(b bool) {
	if s.state != b {
		if *debug {
			fmt.Printf("Changed status of pingentry #%d to %v\n", s.ID, b)

		}
		s.since = time.Now()
	}
	s.state = b
}

func get_status_as_proto(ctx context.Context) []*pb.PingStatus {
	var sts []*status
	stlock.Lock()
	for _, v := range status_trackers {
		sts = append(sts, v)
	}
	stlock.Unlock()

	var res []*pb.PingStatus

	for _, st := range sts {
		pe, err := get_ping_entry_by_id(ctx, st.ID)
		if err != nil {
			fmt.Printf("failed to get entry: %s\n", err)
			continue
		}
		ps := &pb.PingStatus{
			PingEntry: pe,
			Currently: st.state,
			Since:     uint32(st.since.Unix()),
		}
		if pe.IP == "" {
			pe.IP, err = dc.Get(pe.MetricHostName, pe.IPVersion)
			if err != nil {
				fmt.Printf("Failed to resolve: %s\n", err)
			}
		}
		res = append(res, ps)
	}
	return res
}






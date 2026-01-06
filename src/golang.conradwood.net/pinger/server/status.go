package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/prometheus"
)

var (
	stlock          sync.Mutex
	status_trackers = make(map[uint64]*status)
	pingStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pinger_target_status",
			Help: "V=2 U=none DESC=reachable(2) or not(1)",
		},
		[]string{"entryid", "pingerid", "ip", "name", "tag", "tag2", "tag3", "tag4"},
	)
)

type status struct {
	ID           uint64
	state        bool
	since        time.Time
	pingerid     string
	last_updated time.Time // eventually they become stale and are removed
	pe           *pb.PingEntry
}

func init() {
	prometheus.MustRegister(pingStatusGauge)
}
func reset_status_trackers() {
	stlock.Lock()
	status_trackers = make(map[uint64]*status)
	stlock.Unlock()
}
func get_status_tracker(ID uint64, pingerid string) *status {
	var err error
	pe := GetPingEntryRouteByID(ID)
	if pe == nil {
		pe, err = get_ping_entry_by_id(context.Background(), ID)
		if err != nil {
			fmt.Printf("failed to get entry: %s\n", err)
			return nil
		}
	}
	stlock.Lock()
	defer stlock.Unlock()
	res, found := status_trackers[ID]
	if found {
		if res.pe.ID != pe.ID {
			fmt.Printf("In status: %d,%s\n", res.pe.ID, res.pingerid)
			fmt.Printf("In submit: %d,%s\n", ID, pingerid)
			panic("mismatched pingentry id")
		}
		if res.pingerid != pingerid {
			fmt.Printf("In status: %d,%s\n", res.pe.ID, res.pingerid)
			fmt.Printf("In submit: %d,%s\n", ID, pingerid)
			panic("mismatched pingerid")
		}
		return res
	}
	res = &status{ID: ID, since: time.Now(), pingerid: pingerid, pe: pe}
	status_trackers[ID] = res
	return res
}
func (s *status) Set(b bool) {
	s.last_updated = time.Now()
	if s.state != b {
		if *debug {
			fmt.Printf("Changed status of pingentry #%d to %v\n", s.ID, b)

		}
		s.since = time.Now()
	}

	l := s.labels()
	val := 0
	if in_network_status(s.pe) {
		if b {
			val = 2
		} else {
			val = 1
		}
	}
	pingStatusGauge.With(l).Set(float64(val))

	s.state = b

}
func (s *status) labels() prometheus.Labels {
	pe := s.pe

	ip := pe.IP
	if ip == "" {
		nip, err := dc.Get(pe.MetricHostName, pe.IPVersion)
		if err == nil {
			ip = nip
		}
	}

	l := prometheus.Labels{
		"entryid":  fmt.Sprintf("%d", pe.ID),
		"pingerid": s.pingerid,
		"ip":       ip,
		"name":     pe.MetricHostName,
		"tag":      pe.Label,
		"tag2":     pe.Label2,
		"tag3":     pe.Label3,
		"tag4":     pe.Label4,
	}
	return l
}

func get_status_as_proto(ctx context.Context) []*pb.PingStatus {
	var sts []*status
	stlock.Lock()
	var remove []uint64
	for k, v := range status_trackers {
		if time.Since(v.last_updated) > time.Duration(10)*time.Minute {
			remove = append(remove, k) // mark stale one
		} else {
			sts = append(sts, v)
		}
	}
	// delete stale ones
	for _, id := range remove {
		delete(status_trackers, id)
	}
	stlock.Unlock()

	var res []*pb.PingStatus
	var err error
	for _, st := range sts {
		pe := GetPingEntryRouteByID(st.ID)
		if pe == nil {
			pe, err = get_ping_entry_by_id(ctx, st.ID)
			if err != nil {
				fmt.Printf("failed to get entry: %s\n", err)
				continue
			}
		}
		if !pe.IsActive {
			continue
		}
		ps := &pb.PingStatus{
			PingEntry: pe,
			Currently: st.state,
			Since:     uint32(st.since.Unix()),
		}
		// set the 5minute state
		b := true
		if (!st.state) && time.Since(st.since) >= time.Duration(5)*time.Minute {
			b = false
		}
		ps.State5Min = b
		if pe.IP == "" {
			pe.IP, err = dc.Get(pe.MetricHostName, pe.IPVersion)
			if err != nil {
				fmt.Printf("status entryid #%d: Failed to resolve: %s\n", pe.ID, err)
			}
		}

		res = append(res, ps)
	}
	return res
}

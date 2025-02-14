package main

import (
	"sync"
	"time"

	pb "golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/prometheus"
)

var (
	pingstatelock   sync.Mutex
	pingState       = make(map[string]*PingState)
	pingStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pinger_target_status",
			Help: "V=2 U=none DESC=reachable(2) or not(1)",
		},
		[]string{"pingerid", "ip", "name", "tag"},
	)
	pingSpeed = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "pinger_target_speed",
			Help: "V=2 U=none DESC=ping latency in seconds",
		},
		[]string{"pingerid", "ip", "name", "tag"},
	)
)

type PingState struct {
	pe                    *pb.PingEntry
	lastAttempt           time.Time
	last_failed_ping      time.Time
	last_successful_ping  time.Time
	last_latency          time.Duration
	first_successful_ping time.Time
	first_failed_ping     time.Time
	failctr               int
	successctr            int
}

func init() {
	prometheus.MustRegister(pingStatusGauge, pingSpeed)
}
func getAllPingStates() []*PingState {
	var res []*PingState
	pingstatelock.Lock()
	defer pingstatelock.Unlock()
	for _, v := range pingState {
		res = append(res, v)
	}
	return res
}
func getPingState(pe *pb.PingEntry) *PingState {
	pingstatelock.Lock()
	defer pingstatelock.Unlock()
	ps := pingState[pe.IP]
	if ps != nil {
		return ps
	}
	ps = &PingState{pe: pe}
	pingState[pe.IP] = ps
	return ps

}

func (ps *PingState) Failed() {
	now := time.Now()
	ps.last_failed_ping = now
	if ps.failctr == 0 {
		ps.first_failed_ping = now
	}
	ps.failctr++
	ps.successctr = 0
	ps.UpdateGauge()
}
func (ps *PingState) Success(td time.Duration) {
	now := time.Now()
	ps.last_latency = td
	ps.last_successful_ping = now
	if ps.successctr == 0 {
		ps.first_successful_ping = now
	}
	ps.successctr++
	ps.failctr = 0
	ps.UpdateGauge()
}
func (ps *PingState) IsReachable() bool {
	if ps.successctr > 0 {
		return true
	}
	return false
}
func (ps *PingState) UpdateGauge() {
	l := prometheus.Labels{"pingerid": *pingerid,
		"ip":   ps.pe.IP,
		"name": ps.pe.MetricHostName,
		"tag":  ps.pe.Label,
	}
	val := 0
	if ps.successctr > 0 {
		val = 2
	} else if ps.failctr > 0 {
		val = 1
	}
	pingStatusGauge.With(l).Set(float64(val))
	if ps.failctr == 0 {
		pingSpeed.With(l).Observe(ps.last_latency.Seconds())
	}
}
func (ps *PingState) PingTargetStatus() *pb.PingTargetStatus {
	pe := ps.pe
	res := &pb.PingTargetStatus{
		IP:   pe.IP,
		Name: pe.MetricHostName,
	}

	if ps.successctr > 0 {
		res.Since = uint32(ps.first_successful_ping.Unix())
		res.Reachable = true
	} else if ps.failctr > 0 {
		res.Since = uint32(ps.first_failed_ping.Unix())
		res.Reachable = false
	}

	return res
}

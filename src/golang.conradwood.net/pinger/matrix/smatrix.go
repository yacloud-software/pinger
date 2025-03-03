package matrix

import (
	"fmt"
	"sort"

	"golang.conradwood.net/apis/pinger"
)

type smatrix struct {
	st            []*pinger.PingStatus
	known_pingers []string // index equivalent to columns
}

func (sm *smatrix) GetColumnHeadings() []*pinger.MatrixColumnHeading {
	col_headings_map := make(map[string]bool)
	for _, ps := range sm.st {
		key := ps.PingEntry.PingerID
		col_headings_map[key] = true
	}
	var pingers []string
	for k, _ := range col_headings_map {
		pingers = append(pingers, k)
	}
	sort.Slice(pingers, func(i, j int) bool {
		return pingers[i] < pingers[j]
	})
	sm.known_pingers = pingers

	var res []*pinger.MatrixColumnHeading
	for _, p := range sm.known_pingers {
		res = append(res, &pinger.MatrixColumnHeading{DisplayName: p})
	}
	return res
}

func (sm *smatrix) PingerStatusListForHostIP(host, ip string) []*pinger.PingStatus {
	var res []*pinger.PingStatus
	for _, ps := range sm.st {
		if ps.PingEntry.MetricHostName != host {
			continue
		}
		if ps.PingEntry.IP != ip {
			continue
		}
		res = append(res, ps)
	}
	return res

}

type hostip struct {
	sm   *smatrix
	host string
	ip   string
}

func (sm *smatrix) GetAllHostIPCombos() []*hostip {
	resmap := make(map[string]*hostip)
	for _, ps := range sm.st {
		key := pingstatus2key(ps)
		hi := &hostip{sm: sm, host: ps.PingEntry.MetricHostName, ip: ps.PingEntry.IP}
		resmap[key] = hi
	}
	var res []*hostip
	for _, v := range resmap {
		res = append(res, v)
	}
	return res
}

func (h *hostip) GetPingStatusForColumn(col int) (string, *pinger.PingStatus) {
	//	pslist := h.sm.PingerStatusListForHostIP(h.host, h.ip)
	pinger := h.sm.known_pingers[col]
	s := fmt.Sprintf("%s -> %s_%s", pinger, h.host, h.ip)
	return s, nil
}

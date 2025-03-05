package matrix

import "golang.conradwood.net/apis/pinger"

// attempt to identify which networks do not route to which other network(s)
func build_by_network_matrix(st []*pinger.PingStatus) (*pinger.StatusMatrix, error) {
	// filter out any that do not have netroutes
	var vst []*pinger.PingStatus
	for _, ps := range st {
		if ps.PingEntry.NetRouteConfig == nil {
			continue
		}
		vst = append(vst, ps)
	}

	res := &pinger.StatusMatrix{}
	nl := &networklist{}
	for _, ps := range vst {
		nl.AddIP(ps.PingEntry.IP)
	}
	return res, nil
}

package dot

import (
	pb "golang.conradwood.net/apis/pinger"
)

func GenerateDotFromPingStatus(status *pb.PingStatusList) (string, error) {
	nodes := make(map[string]*node)
	for _, ps := range status.Status {
		n := nodes[ps.PingEntry.PingerID]
		if n == nil {
			n = &node{Name: ps.PingEntry.PingerID}
			nodes[ps.PingEntry.PingerID] = n
		}
		nt := &nodetarget{
			Name: ps.PingEntry.MetricHostName,
			ok:   ps.Currently,
		}
		n.Targets = append(n.Targets, nt)
	}
	var nodea []*node
	for _, v := range nodes {
		nodea = append(nodea, v)
	}
	return GenerateDot(nodea)
}

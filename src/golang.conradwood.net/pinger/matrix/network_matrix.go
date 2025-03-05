package matrix

import (
	"fmt"
	"strings"

	"golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/utils"
)

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
	fmt.Printf("   network matrix for %d pingstatus of which %d have netroutes\n", len(st), len(vst))

	res := &pinger.StatusMatrix{}
	nl := &networklist{}
	for _, ps := range vst {
		nrc := ps.PingEntry.NetRouteConfig
		route := nrc.Route
		fmt.Printf("   Pingentry %s -> %s (%s)\n", ps.PingEntry.PingerID, ps.PingEntry.MetricHostName, ps.PingEntry.IP)
		for _, ip := range nrc.ToHost.IPs {
			_, _, v, err := utils.ParseIP(ip)
			if err != nil {
				fmt.Printf("parse ip error: %s\n", err)
				continue
			}
			if uint32(v) != route.IPVersion {
				continue
			}
			//fmt.Printf("      %v\n", ip)
			if strings.HasPrefix(ip, ps.PingEntry.IP) {
				nl.AddIP(nrc.ToHost, ip) // networkdef of IP
			}
		}
	}
	for _, ps := range vst {
		nrc := ps.PingEntry.NetRouteConfig
		from_ip := find_source_ip_for_dest(nrc.FromHost, ps.PingEntry.IP)
		from_net_info := lookup_net_info(from_ip)
		to_net_info := lookup_net_info(ps.PingEntry.IP)
		nl.Record(from_net_info.asn, to_net_info.asn, ps.Currently)
	}

	t := utils.Table{}
	for _, nets := range nl.GetNetworks() {
		t.AddString(nets.asn)
		for _, f := range nets.failures {
			t.AddString(f.asn)
		}
		t.NewRow()
	}
	fmt.Println(t.ToPrettyString())
	t = utils.Table{}
	for _, nets := range nl.GetNetworks() {
		t.AddString(nets.asn)
		for _, f := range nets.successes {
			t.AddString(f.asn)
		}
		t.NewRow()
	}
	fmt.Println(t.ToPrettyString())

	return res, nil
}

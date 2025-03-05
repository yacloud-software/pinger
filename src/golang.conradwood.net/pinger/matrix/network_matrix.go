package matrix

import (
	"fmt"
	"strings"

	"golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/utils"
)

type net_matrix struct {
}

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

	m := &amatrix{}
	for _, net := range nl.GetNetworks() {
		for _, r := range net.records {
			m.SetCellByName(r.to_network.asn, net.asn, &netstatus{ctr: r})
		}
	}

	for _, s := range m.GetColumnNames() {
		res.ColumnHeadings = append(res.ColumnHeadings, &pinger.MatrixColumnHeading{DisplayName: s})
	}
	for _, row := range m.Rows() {
		mrow := &pinger.MatrixRow{Hostname: row.Name()}
		res.Rows = append(res.Rows, mrow)
		for _, c := range row.Cells() {
			s := ""
			colour := ""
			no := c.Content()
			if no != nil {
				n := no.(*netstatus)
				s = n.String()
				colour = n.Colour()
			}
			me := &pinger.MatrixEntry{
				DisplayName:   s,
				DisplayColour: colour,
			}
			mrow.Entries = append(mrow.Entries, me)
		}
	}

	return res, nil
}

type netstatus struct {
	ctr *networkdef_ctr
}

func (n *netstatus) String() string {
	if n == nil {
		return ""
	}
	status := "OK"
	r := n.ctr
	if r.failure_count > 0 {
		if r.success_count > 0 {
			status = "PARTIAL"
		} else {
			status = "FAIL"
		}
	}

	return status
}
func (n *netstatus) Colour() string {
	if n == nil {
		return ""
	}
	status := COLOUR_GOOD
	r := n.ctr
	if r.failure_count > 0 {
		if r.success_count > 0 {
			status = COLOUR_WARN
		} else {
			status = COLOUR_FAIL
		}
	}

	return status
}

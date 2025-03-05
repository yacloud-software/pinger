package matrix

import (
	"fmt"
	"net"
	"sync"

	"golang.conradwood.net/apis/geoip"
	"golang.conradwood.net/apis/netroutes"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
)

type networklist struct {
	sync.Mutex
	networks []*networkdef
}
type networkdef struct {
	sync.Mutex
	asn     string
	network string
	hosts   []string
	records []*networkdef_ctr // results between two networks

}
type networkdef_ctr struct {
	to_network    *networkdef
	success_count int
	failure_count int
}
type netinfo struct {
	asn string
	isp string
}

func (nl *networklist) AddIP(host *netroutes.Host, ip string) *networkdef {
	nl.Lock()
	defer nl.Unlock()
	nd := nl.getNetByIP(ip)
	if nd != nil {
		return nd
	}
	ni := lookup_net_info(ip)
	fmt.Printf("        Adding: %s  asn=\"%s\", isp=%s\n", ip, ni.asn, ni.isp)
	nd = &networkdef{asn: ni.asn}
	nl.networks = append(nl.networks, nd)
	return nd

}
func (nl *networklist) getNetByASN(asn string) *networkdef {
	for _, network := range nl.networks {
		if network.asn == asn {
			return network
		}
	}
	return nil
}
func (nl *networklist) getNetByIP(ip string) *networkdef {
	ni := lookup_net_info(ip)
	for _, network := range nl.networks {
		if network.asn == ni.asn {
			return network
		}
	}
	return nil
}
func (nl *networklist) GetNetworks() []*networkdef {
	return nl.networks
}

func (nl *networklist) Record(from_asn, to_asn string, success bool) {
	if from_asn == to_asn {
		return
	}
	nl.Lock()
	defer nl.Unlock()
	from_net := nl.getNetByASN(from_asn)
	if from_net == nil {
		fmt.Printf("no from-net for \"%s\"\n", from_asn)
		return
	}
	to_net := nl.getNetByASN(to_asn)
	if to_net == nil {
		fmt.Printf("no to-net for \"%s\"\n", to_asn)
		return
	}
	// find therecord, add it if necessary
	record := from_net.findRecordForASN(nl, to_asn)
	if success {
		record.success_count++
	} else {
		record.failure_count++
	}
}
func (nd *networkdef) findRecordForASN(nl *networklist, asn string) *networkdef_ctr {
	for _, r := range nd.records {
		if r.to_network.asn == asn {
			return r
		}
	}
	asn_net := nl.getNetByASN(asn)
	r := &networkdef_ctr{
		to_network: asn_net,
	}
	nd.records = append(nd.records, r)
	return r
}

func lookup_net_info(ip string) *netinfo {
	res := &netinfo{}
	b, err := utils.IsPrivateIP(ip)
	if err != nil {
		return res
	}
	if b {
		return lookup_private_net_info(ip)
	}

	pips := ip
	pip, _, err := net.ParseCIDR(ip)
	if err == nil {
		pips = fmt.Sprintf("%v", pip)
	}

	lr, err := geoip.GetGeoIPClient().Lookup(authremote.Context(), &geoip.LookupRequest{IP: pips})
	if err != nil {
		fmt.Printf("No geoip lookup (%s)", err)
		return res
	}

	isp := lr.ISP
	if isp == "" {
		isp = lr.Organisation
	}
	res.isp = isp
	res.asn = lr.AS
	return res
}
func lookup_private_net_info(ip string) *netinfo {
	res := &netinfo{}

	_, ipnet, err := net.ParseCIDR(ip)
	if ipnet == nil || err != nil {
		_, ipnet, err = net.ParseCIDR(ip + "/24")
	}
	if ipnet == nil {
		res.asn = fmt.Sprintf("ASN_LOCAL_%v", ipnet)
		res.isp = "local"
		return res
	}

	res.asn = fmt.Sprintf("ASN_LOCAL_%v", ipnet)
	res.isp = "local"
	return res
}

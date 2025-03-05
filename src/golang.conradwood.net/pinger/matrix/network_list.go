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
	asn       string
	network   string
	hosts     []string
	successes []*networkdef // success to his net
	failures  []*networkdef // failure to his net
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
	if success {
		from_net.AddSuccess(to_net)
	} else {
		from_net.AddFailure(to_net)
	}

}

// record failure from this network to that ip
func (nd *networkdef) AddFailure(ip_net *networkdef) {
	nd.Lock()
	defer nd.Unlock()
	for _, n := range nd.failures {
		if n.asn == ip_net.asn {
			return
		}
	}
	nd.failures = append(nd.failures, ip_net)
}

// record success from this network to that ip
func (nd *networkdef) AddSuccess(ip_net *networkdef) {
	nd.Lock()
	defer nd.Unlock()
	for _, n := range nd.successes {
		if n.asn == ip_net.asn {
			return
		}
	}
	nd.successes = append(nd.successes, ip_net)
}

func lookup_net_info(ip string) *netinfo {
	res := &netinfo{}
	pips := ip
	pip, net, err := net.ParseCIDR(ip)
	if err == nil {
		pips = fmt.Sprintf("%v", pip)
	}
	//	net_ip := fmt.Sprintf("%v", net.IP)
	//	net_size, _ := net.Mask.Size()
	b, err := utils.IsPrivateIP(ip)
	if err != nil {
		return res
	}
	if b {
		res.asn = fmt.Sprintf("ASN_LOCAL_%v", net)
		res.isp = "local"
		return res
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

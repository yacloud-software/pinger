package matrix

import (
	"fmt"

	"golang.conradwood.net/apis/netroutes"
	"golang.conradwood.net/go-easyops/utils"
)

// given a host (with multiple ips) and a destination IP, work out which source ip it is likely to use
func find_source_ip_for_dest(host *netroutes.Host, to_ip string) string {
	// really we net the routes on that host to determine this.
	// this is a wild guess at best
	b, err := utils.IsPrivateIP(to_ip)
	if err != nil {
		return ""
	}
	if b {
		// do not do private ips yet
		return ""
	}
	_, _, to_ver, err := utils.ParseIP(to_ip)
	if err != nil {
		fmt.Printf("invalid ip: %s\n", err)
		return ""
	}
	for _, host_ip := range host.IPs {
		b, err := utils.IsPrivateIP(host_ip)
		if err != nil {
			return ""
		}
		if b {
			// do not do private ips yet
			continue
		}

		_, _, from_ver, err := utils.ParseIP(host_ip)
		if err != nil {
			fmt.Printf("invalid host ip: %s\n", err)
			continue
		}
		if from_ver != to_ver {
			// do not consider 4to6 gateways etc
			continue
		}
		return host_ip

	}
	return ""
}

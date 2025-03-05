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
	to_is_private, err := utils.IsPrivateIP(to_ip)
	if err != nil {
		return ""
	}

	_, _, to_ver, err := utils.ParseIP(to_ip)
	if err != nil {
		fmt.Printf("invalid ip: %s\n", err)
		return ""
	}

	// try first with non /32
	for _, host_ip := range host.IPs {
		hostip_is_private, err := utils.IsPrivateIP(host_ip)
		if err != nil {
			return ""
		}
		if hostip_is_private != to_is_private {
			// ignoring NAT gateways...
			continue
		}

		_, netsize, from_ver, err := utils.ParseIP(host_ip)
		if err != nil {
			fmt.Printf("invalid host ip: %s\n", err)
			continue
		}
		if netsize == 32 {
			continue
		}
		if from_ver != to_ver {
			// do not consider 4to6 gateways etc
			continue
		}
		return host_ip
	}
	// try any
	for _, host_ip := range host.IPs {
		hostip_is_private, err := utils.IsPrivateIP(host_ip)
		if err != nil {
			return ""
		}
		if hostip_is_private != to_is_private {
			// ignoring NAT gateways...
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

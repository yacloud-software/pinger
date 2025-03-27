package main

import (
	"fmt"
	"sync"

	"golang.conradwood.net/apis/netroutes"
	"golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/authremote"
)

var (
	route_id     = 1000
	route_lock   sync.Mutex
	known_routes = make(map[string]*pinger.PingEntry)
	ignore_ips   = map[string][]string{
		"scweb": []string{"172.29.2.254"},
	}
)

func GetPingEntryRouteByID(id uint64) *pinger.PingEntry {
	route_lock.Lock()
	defer route_lock.Unlock()
	for _, v := range known_routes {
		if v.ID == id {
			return v
		}
	}
	return nil
}
func GetRoutesFromNetRoutes(fromhost string) ([]*pinger.PingEntry, error) {
	route_lock.Lock()
	defer route_lock.Unlock()
	req := &netroutes.RouteRequest{
		IncludePublic:  true,
		IncludePrivate: true,
	}
	routes, err := netroutes.GetNetRoutesClient().GetRoutes(authremote.Context(), req)
	if err != nil {
		return nil, err
	}
	var res []*pinger.PingEntry
	for _, route := range routes.Routes {
		rfromhost := routes.Hosts[route.FromHost]
		rtohost := routes.Hosts[route.ToHost]
		if rfromhost == nil {
			fmt.Printf("NO fromhost for ID #%d\n", route.FromHost)
			continue
		}
		if rtohost == nil {
			fmt.Printf("NO tohost for ID #%d\n", route.ToHost)
			continue
		}
		if rfromhost.Name != fromhost {
			continue
		}
		if ignore_ip(rtohost.Name, route.ToIP) {
			continue
		}

		key := fmt.Sprintf("%d_%s", route.FromHost, route.ToIP)
		pe := known_routes[key]
		if pe == nil {
			pe = &pinger.PingEntry{
				ID:             500,
				IP:             route.ToIP,
				Interval:       10,
				MetricHostName: rtohost.Name,
				PingerID:       rfromhost.Name,
				IPVersion:      route.IPVersion,
				IsActive:       true,
				NetRouteConfig: &pinger.NetRouteConfig{FromHost: rfromhost, ToHost: rtohost, Route: route},
			}
			pe.ID = uint64(route_id)
			route_id++
			known_routes[key] = pe
		}
		res = append(res, pe)
	}
	return res, nil
}

func ignore_ip(name, ip string) bool {
	//	fmt.Printf("Ignore ip \"%s\" on host \"%s\"?\n", ip, name)
	ignips := ignore_ips[name]
	for _, i := range ignips {
		if i == ip {
			return true
		}
	}
	return false
}

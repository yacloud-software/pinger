package main

import (
	"fmt"
	"sync"

	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/apis/netroutes"
	"golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/authremote"
)

var (
	route_id     = 1000
	route_lock   sync.Mutex
	known_routes = make(map[string]*pinger.PingEntry)
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
	routes, err := netroutes.GetNetRoutesClient().GetRoutes(authremote.Context(), &common.Void{})
	if err != nil {
		return nil, err
	}
	var res []*pinger.PingEntry
	for _, route := range routes.Routes {
		if route.FromHost != fromhost {
			continue
		}
		key := fmt.Sprintf("%s_%s", route.FromHost, route.ToIP)
		pe := known_routes[key]
		if pe == nil {
			pe = &pinger.PingEntry{
				ID:             500,
				IP:             route.ToIP,
				Interval:       10,
				MetricHostName: route.ToHost,
				PingerID:       route.FromHost,
				IPVersion:      route.IPVersion,
				IsActive:       true,
			}
			pe.ID = uint64(route_id)
			route_id++
			known_routes[key] = pe
		}
		res = append(res, pe)
	}
	return res, nil
}

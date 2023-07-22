package main

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"
)

type DNSCache struct {
	entries []*dnsCacheEntry
}

type dnsCacheEntry struct {
	name    string
	ipv     uint32
	created time.Time
	ip      string
}

// lookup name or get from cache
func (d *DNSCache) Get(name string, version uint32) (string, error) {
	for _, e := range d.entries {
		if e.name == name && e.ipv == version {
			return e.ip, nil
		}
	}
	ips, err := net.LookupHost(name)
	if err != nil {
		return "", err
	}
	sort.Slice(ips, func(i, j int) bool {
		return ips[i] < ips[j]
	})
	for _, ip := range ips {
		if IPVersion(ip) == version {
			fmt.Printf("Resolved %s (%d) to IP: %#v\n", name, version, ip)
			dce := &dnsCacheEntry{
				name:    name,
				ipv:     version,
				created: time.Now(),
				ip:      ip,
			}
			d.entries = append(d.entries, dce)
			return ip, nil
		}
	}
	fmt.Printf("Failed to resolve %s (%d) to any ip\n", name, version)
	return "", nil
}

func IPVersion(ip string) uint32 {
	if strings.Index(ip, ".") != -1 {
		return 4
	}
	if strings.Index(ip, ":") != -1 {
		return 6
	}
	return 0
}

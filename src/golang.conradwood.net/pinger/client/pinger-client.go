package main

import (
	"flag"
	"fmt"
	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"sync"
)

var (
	echoClient pb.PingerClient
	pinger     = flag.String("pinger", "", "if set query that particular pinger")
	iplist     = flag.Bool("iplist", false, "if true get list of ips")
)

func main() {
	flag.Parse()
	if *iplist {
		utils.Bail("failed to get ipList()", ipList())
		os.Exit(0)
	}
	ips := flag.Args()
	if len(ips) == 0 {
		utils.Bail("failed to get status", Status())
		os.Exit(0)
	}

	echoClient = pb.GetPingerClient()

	// a context with authentication
	ctx := authremote.Context()
	var wg sync.WaitGroup
	for _, lip := range ips {
		wg.Add(1)
		go func(ip string) {
			empty := &pb.PingRequest{IP: ip}
			response, err := echoClient.Ping(ctx, empty)
			utils.Bail("Failed to ping server", err)
			fmt.Printf("Ping to %s: %v (%dms)\n", response.IP, response.Success, response.Milliseconds)
			wg.Done()
		}(lip)
	}
	wg.Wait()

	fmt.Printf("Done.\n")
	os.Exit(0)
}
func ipList() error {
	pl := pb.GetPingerListClient()
	ctx := authremote.Context()
	preq := &pb.PingListRequest{PingerID: *pinger}
	pr, err := pl.GetPingList(ctx, preq)
	if err != nil {
		return err
	}
	t := &utils.Table{}
	t.AddHeaders("metrichost", "ipv", "ip", "label")
	for _, e := range pr.Entries {
		t.AddString(e.MetricHostName)
		t.AddUint32(e.IPVersion)
		t.AddString(e.IP)
		t.AddString(e.Label)
		t.NewRow()
	}
	fmt.Println(t.ToPrettyString())
	return nil
}
func Status() error {
	ctx := authremote.Context()
	if *pinger != "" {
		ctx = authremote.DerivedContextWithRouting(ctx, map[string]string{"pinger": *pinger}, false)
	}
	res, err := pb.GetPingerClient().PingStatus(ctx, &common.Void{})
	if err != nil {
		return err
	}
	fmt.Printf("Got %d ping status\n", len(res.Status))
	t := utils.Table{}
	t.AddHeaders("IP", "Name", "Reachable", "Since")
	for _, ps := range res.Status {
		t.AddString(ps.IP)
		t.AddString(ps.Name)
		t.AddBool(ps.Reachable)
		t.AddTimestamp(ps.Since)
		t.NewRow()
	}
	fmt.Println(t.ToPrettyString())
	return nil
}

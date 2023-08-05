package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-ping/ping"
	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/server"
	"golang.conradwood.net/go-easyops/utils"
	"golang.singingcat.net/scgolib/goodness"
	"google.golang.org/grpc"
	"os"
	"time"
)

var (
	debug    = flag.Bool("debug", false, "debug mode")
	port     = flag.Int("port", 4107, "The grpc server port")
	pingerid = flag.String("pingerid", "", "pingerid")
	pinglist *pb.PingList
	gn       goodness.Goodness
)

type echoServer struct {
}

func main() {
	flag.Parse()
	fmt.Printf("Starting PingerServer...\n")

	gn = goodness.NewGoodness("ping")
	go pinger_list_update_loop()
	go pingstuff_loop()
	sd := server.NewServerDef()
	sd.AddTag("pinger", *pingerid)
	sd.Port = *port
	sd.Register = server.Register(
		func(server *grpc.Server) error {
			e := new(echoServer)
			pb.RegisterPingerServer(server, e)
			return nil
		},
	)
	err := server.ServerStartup(sd)
	utils.Bail("Unable to start server", err)
	os.Exit(0)
}

/************************************
* ping stuff in list
************************************/
func pingstuff_loop() {
	for {
		time.Sleep(1 * time.Second)
		pingstuff()
	}
}
func pingstuff() {
	if pinglist == nil {
		return
	}
	for _, p := range pinglist.Entries {
		ps := getPingState(p)
		if time.Since(ps.lastAttempt) < time.Duration(ps.pe.Interval)*time.Second {
			continue
		}
		now := time.Now()
		ps.lastAttempt = now
		prefix := fmt.Sprintf("[pinglist %s] ", ps.pe.IP)
		pr, err := singlePing(prefix, ps.pe.IP)
		if err != nil {
			reportStateUpstream(ps, false)
			ps.Failed()
			fmt.Printf("%sFailed to ping %s: %s\n", prefix, ps.pe.IP, err)
			continue
		}
		if pr != nil {
			reportStateUpstream(ps, pr.Success)
			if pr.Success {
				ps.Success()
			} else {
				ps.Failed()
			}
		}
		debugf("%sPinged %s - %v\n", prefix, pr.IP, pr.Success)
	}
}
func reportStateUpstream(ps *PingState, result bool) {
	pe := ps.pe
	fmt.Printf("Reporting upstream: %d (%v)\n", pe.ID, result)
	ctx := authremote.Context()
	r := &pb.SetPingStatusRequest{
		ID:      pe.ID,
		Success: result,
	}
	_, err := pb.GetPingerListClient().SetPingStatus(ctx, r)
	if err != nil {
		fmt.Printf("Failed to inform pingerlist of new status: %s\n", err)
	}

}

/************************************
* update loop
************************************/
func pinger_list_update_loop() {
	time.Sleep(time.Duration(2) * time.Second)
	for {
		pinger_list_update()
		time.Sleep(30 * time.Second)
	}
}
func pinger_list_update() {
	if *pingerid == "" {
		return
	}
	if pinglist == nil || len(pinglist.Entries) == 0 {
		fmt.Printf("Got no pinglist, getting new list from pingerlist service\n")
	}
	ctx := authremote.Context()
	pl, err := pb.GetPingerListClient().GetPingList(ctx, &pb.PingListRequest{PingerID: *pingerid})
	if err != nil {
		fmt.Printf("Failed to get pinglist: %s\n", utils.ErrorString(err))
		return
	}
	op := 0
	np := 0
	if pinglist != nil {
		op = len(pinglist.Entries)
	}
	if pl != nil {
		np = len(pl.Entries)
	}
	if op == 0 || op != np {
		fmt.Printf("Had %d ping targets, now got %d\n", op, np)
	}
	res := &pb.PingList{}
	for _, e := range pl.Entries {
		if e.IP == "" {
			fmt.Printf("Ignoring Entry #%d (%s) - got no ip\n", e.ID, e.MetricHostName)
			continue
		}
		res.Entries = append(res.Entries, e)
	}

	pinglist = res
}

/************************************
* grpc functions
************************************/

func (e *echoServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResult, error) {
	prefix := fmt.Sprintf("[pingrequest %s] ", req.IP)
	return singlePing(prefix, req.IP)
}
func singlePing(prefix string, ip string) (*pb.PingResult, error) {
	debugf("%sPinging %s\n", prefix, ip)
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return nil, err
	}
	pinger.SetPrivileged(true)
	pinger.Count = 1
	pinger.Timeout = time.Duration(3) * time.Second
	started := time.Now()
	err = pinger.Run() // Blocks until finished.
	dur := time.Since(started)
	if err != nil {
		return nil, err
	}
	stats := pinger.Statistics() // get send/receive/rtt stats
	res := &pb.PingResult{IP: ip}
	if stats.PacketsRecv > 0 {
		res.Success = true
	}
	res.Milliseconds = uint32(dur.Milliseconds())
	return res, nil

}
func (e *echoServer) PingStatus(ctx context.Context, req *common.Void) (*pb.PingTargetStatusList, error) {
	res := &pb.PingTargetStatusList{}
	psl := getAllPingStates()
	for _, ps := range psl {
		res.Status = append(res.Status, ps.PingTargetStatus())
	}
	return res, nil
}

func debugf(format string, args ...interface{}) {
	if !*debug {
		return
	}
	fmt.Printf(format, args...)
}

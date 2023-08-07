package main

import (
	"context"
	"flag"
	"fmt"
	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/server"
	"sort"
	//	"golang.conradwood.net/go-easyops/sql"
	"golang.conradwood.net/go-easyops/utils"
	"golang.conradwood.net/pinger/db"
	"golang.singingcat.net/scgolib/goodness"
	"google.golang.org/grpc"
	"os"
)

var (
	dc       = &DNSCache{}
	port     = flag.Int("port", 4100, "The grpc server port")
	pinglist *pb.PingList
	gn       goodness.Goodness
	pedb     *db.DBPingEntry
	//	psql     *sql.DB
	debug = flag.Bool("debug", false, "debug mode")
)

type echoServer struct {
}

func main() {
	flag.Parse()
	fmt.Printf("Starting PingerServer...\n")
	var err error
	//	psql, err = sql.Open()
	utils.Bail("failed to open psql", err)
	pedb = db.DefaultDBPingEntry()
	db.DefaultDBHost()
	db.DefaultDBIP()
	db.DefaultDBRoute()
	db.DefaultDBTag()
	gn = goodness.NewGoodness("ping")
	sd := server.NewServerDef()
	sd.Port = *port
	sd.Register = server.Register(
		func(server *grpc.Server) error {
			e := new(echoServer)
			pb.RegisterPingerListServer(server, e)
			return nil
		},
	)
	err = server.ServerStartup(sd)
	utils.Bail("Unable to start server", err)
	os.Exit(0)
}

/************************************
* grpc functions for the ping manager..
************************************/

func (e *echoServer) GetPingList(ctx context.Context, req *pb.PingListRequest) (*pb.PingList, error) {
	if *debug {
		fmt.Printf("Requested pinglist for pingerid: \"%s\"\n", req.PingerID)
	}
	if req.PingerID == "" {
		return nil, errors.InvalidArgs(ctx, "invalid pingerid", "invalid pingerid")
	}
	ape, err := pedb.ByPingerID(ctx, req.PingerID)
	if err != nil {
		return nil, err
	}
	res := &pb.PingList{}

	var entries []*pb.PingEntry
	for _, e := range ape {
		if e.IsActive == false {
			continue
		}
		entries = append(entries, e)
		if e.IP == "" {
			e.IP, err = dc.Get(e.MetricHostName, e.IPVersion)
			if err != nil {
				fmt.Printf("Failed to lookup %s: %s\n", e.MetricHostName, err)

			}
		}
	}
	res.Entries = entries
	if *debug {
		fmt.Printf("Returned %d entries to pinger %s\n", len(res.Entries), req.PingerID)
	}
	return res, nil
}
func (e *echoServer) SetPingStatus(ctx context.Context, req *pb.SetPingStatusRequest) (*common.Void, error) {
	get_status_tracker(req.ID).Set(req.Success)
	return &common.Void{}, nil
}
func (e *echoServer) GetPingStatus(ctx context.Context, req *common.Void) (*pb.PingStatusList, error) {
	res := &pb.PingStatusList{
		Status: get_status_as_proto(ctx),
	}
	sort.Slice(res.Status, func(i, j int) bool {
		s1 := res.Status[i].PingEntry
		s2 := res.Status[j].PingEntry
		if s1.PingerID != s2.PingerID {
			return s1.PingerID < s2.PingerID
		}
		return s1.MetricHostName < s2.MetricHostName
	})
	return res, nil
}

func get_ping_entry_by_id(ctx context.Context, ID uint64) (*pb.PingEntry, error) {
	r, err := pedb.ByID(ctx, ID)
	return r, err
}

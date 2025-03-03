package main

import (
	"context"
	"flag"
	"fmt"
	"sort"

	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/apis/pinger"
	pb "golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/server"

	//	"golang.conradwood.net/go-easyops/sql"
	"os"

	"golang.conradwood.net/go-easyops/utils"
	"golang.conradwood.net/pinger/db"
	"golang.singingcat.net/scgolib/goodness"
	"google.golang.org/grpc"
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
	sd.SetPort(*port)
	sd.SetRegister(server.Register(
		func(server *grpc.Server) error {
			e := new(echoServer)
			pb.RegisterPingerListServer(server, e)
			return nil
		},
	))
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

	entries, err := GetRoutesFromNetRoutes(req.PingerID)
	if err != nil {
		fmt.Printf("failed to get netroutes for \"%s\"\n", err)
	}
	for _, e := range ape {
		if e.IsActive == false {
			continue
		}
		// exists already?
		if find_entry(entries, e) != nil {
			continue
		}
		if e.IP == "" {
			e.IP, err = dc.Get(e.MetricHostName, e.IPVersion)
			if err != nil {
				fmt.Printf("Failed to lookup %s: %s\n", e.MetricHostName, err)

			}
		}
		entries = append(entries, e)

	}
	res.Entries = entries

	if *debug {
		fmt.Printf("Returned %d entries to pinger %s\n", len(res.Entries), req.PingerID)
	}
	return res, nil
}
func (e *echoServer) SetPingStatus(ctx context.Context, req *pb.SetPingStatusRequest) (*common.Void, error) {
	st := get_status_tracker(req.ID)
	if st == nil {
		fmt.Printf("Submitted Status #%d from \"%s\" not valid\n", req.ID, req.PingerID)
		return nil, errors.InvalidArgs(ctx, "invalid id", "invalid id %d", req.ID)
	}
	st.Set(req.Success)
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

func find_entry(entries []*pinger.PingEntry, e *pinger.PingEntry) *pinger.PingEntry {
	for _, ex := range entries {
		if ex.IP != e.IP {
			continue
		}
		if ex.IPVersion != e.IPVersion {
			continue
		}
		if ex.MetricHostName != e.MetricHostName {
			continue
		}
		if ex.PingerID != e.PingerID {
			continue
		}
		return ex
	}
	return nil
}

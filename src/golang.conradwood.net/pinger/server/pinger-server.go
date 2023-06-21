package main

import (
	"context"
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/server"
	"golang.conradwood.net/go-easyops/sql"
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
	psql     *sql.DB
	debug    = flag.Bool("debug", false, "debug mode")
)

type echoServer struct {
}

func main() {
	flag.Parse()
	fmt.Printf("Starting PingerServer...\n")
	var err error
	psql, err = sql.Open()
	utils.Bail("failed to open psql", err)
	pedb = db.NewDBPingEntry(psql)
	gn = goodness.NewGoodness("ping")
	go ping_loop()
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
	res := &pb.PingList{
		Entries: ape,
	}
	for _, e := range res.Entries {
		if e.IP == "" {
			e.IP, err = dc.Get(e.MetricHostName, e.IPVersion)
			if err != nil {
				fmt.Printf("Failed to lookup %s: %s\n", e.MetricHostName, err)

			}
		}
	}
	return res, nil
}

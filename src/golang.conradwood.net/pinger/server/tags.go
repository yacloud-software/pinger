package main

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/pinger/db"
)

func (e *echoServer) AddTag(ctx context.Context, req *pb.AddTagRequest) (*pb.TagList, error) {
	return &pb.TagList{}, nil
}
func (e *echoServer) RemoveTag(ctx context.Context, req *pb.RemoveTagRequest) (*pb.TagList, error) {
	return &pb.TagList{}, nil
}
func (e *echoServer) CreateRoute(ctx context.Context, req *pb.Route) (*pb.Route, error) {
	return req, nil
}
func (e *echoServer) CreateHost(ctx context.Context, req *pb.Host) (*pb.Host, error) {
	_, err := db.DefaultDBHost().Save(ctx, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}
func (e *echoServer) AddIP(ctx context.Context, req *pb.AddIPRequest) (*pb.IP, error) {
	if req.Name == "" && req.IP == "" {
		return nil, errors.InvalidArgs(ctx, "missing name and ip", "missing name and ip")
	}
	if req.HostID == 0 {
		return nil, errors.InvalidArgs(ctx, "missing hostid", "missing hostid")
	}
	ipv := req.IPVersion
	if req.IP != "" {
		ipv = IPVersion(req.IP)
		if ipv == 0 {
			return nil, errors.InvalidArgs(ctx, "invalid ip", "invalid ip (%s)", req.IP)
		}
		if req.IPVersion != 0 && ipv != req.IPVersion {
			return nil, errors.InvalidArgs(ctx, "ip version mismatch", " ip \"%s\" is an IPv%d address, but requested was was %d)", req.IP, ipv, req.IPVersion)
		}
	}
	if ipv == 0 {
		return nil, errors.InvalidArgs(ctx, "missing ipversion", "missing ipversion")
	}
	if ipv != 6 && ipv != 4 {
		return nil, errors.InvalidArgs(ctx, "ipversion must be either 4 or 6", "weird ipversion %d", ipv)
	}
	fmt.Printf("Adding version %d, name=\"%s\", ip=\"%s\"\n", req.IPVersion, req.Name, req.IP)
	ip := &pb.IP{
		Host:      &pb.Host{ID: req.HostID},
		Name:      req.Name,
		IPVersion: ipv,
		IP:        req.IP,
	}
	_, err := db.DefaultDBIP().Save(ctx, ip)
	if err != nil {
		return nil, err
	}
	return ip, nil
}


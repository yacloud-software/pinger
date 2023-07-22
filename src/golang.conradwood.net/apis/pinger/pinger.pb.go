// Code generated by protoc-gen-go.
// source: protos/golang.conradwood.net/apis/pinger/pinger.proto
// DO NOT EDIT!

/*
Package pinger is a generated protocol buffer package.

It is generated from these files:
	protos/golang.conradwood.net/apis/pinger/pinger.proto

It has these top-level messages:
	PingResult
	PingRequest
	PingListRequest
	PingList
	PingEntry
	PingTargetStatus
	PingTargetStatusList
	SetPingStatusRequest
	PingStatus
	PingStatusList
*/
package pinger

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import common "golang.conradwood.net/apis/common"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// result of a single ping
type PingResult struct {
	IP           string `protobuf:"bytes,1,opt,name=IP" json:"IP,omitempty"`
	Milliseconds uint32 `protobuf:"varint,2,opt,name=Milliseconds" json:"Milliseconds,omitempty"`
	Success      bool   `protobuf:"varint,3,opt,name=Success" json:"Success,omitempty"`
}

func (m *PingResult) Reset()                    { *m = PingResult{} }
func (m *PingResult) String() string            { return proto.CompactTextString(m) }
func (*PingResult) ProtoMessage()               {}
func (*PingResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *PingResult) GetIP() string {
	if m != nil {
		return m.IP
	}
	return ""
}

func (m *PingResult) GetMilliseconds() uint32 {
	if m != nil {
		return m.Milliseconds
	}
	return 0
}

func (m *PingResult) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

// ping something once
type PingRequest struct {
	IP string `protobuf:"bytes,1,opt,name=IP" json:"IP,omitempty"`
}

func (m *PingRequest) Reset()                    { *m = PingRequest{} }
func (m *PingRequest) String() string            { return proto.CompactTextString(m) }
func (*PingRequest) ProtoMessage()               {}
func (*PingRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *PingRequest) GetIP() string {
	if m != nil {
		return m.IP
	}
	return ""
}

type PingListRequest struct {
	PingerID string `protobuf:"bytes,1,opt,name=PingerID" json:"PingerID,omitempty"`
}

func (m *PingListRequest) Reset()                    { *m = PingListRequest{} }
func (m *PingListRequest) String() string            { return proto.CompactTextString(m) }
func (*PingListRequest) ProtoMessage()               {}
func (*PingListRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PingListRequest) GetPingerID() string {
	if m != nil {
		return m.PingerID
	}
	return ""
}

type PingList struct {
	Entries []*PingEntry `protobuf:"bytes,1,rep,name=Entries" json:"Entries,omitempty"`
}

func (m *PingList) Reset()                    { *m = PingList{} }
func (m *PingList) String() string            { return proto.CompactTextString(m) }
func (*PingList) ProtoMessage()               {}
func (*PingList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *PingList) GetEntries() []*PingEntry {
	if m != nil {
		return m.Entries
	}
	return nil
}

type PingEntry struct {
	ID             uint64 `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	IP             string `protobuf:"bytes,2,opt,name=IP" json:"IP,omitempty"`
	Interval       uint32 `protobuf:"varint,3,opt,name=Interval" json:"Interval,omitempty"`
	MetricHostName string `protobuf:"bytes,4,opt,name=MetricHostName" json:"MetricHostName,omitempty"`
	PingerID       string `protobuf:"bytes,5,opt,name=PingerID" json:"PingerID,omitempty"`
	Label          string `protobuf:"bytes,6,opt,name=Label" json:"Label,omitempty"`
	IPVersion      uint32 `protobuf:"varint,7,opt,name=IPVersion" json:"IPVersion,omitempty"`
}

func (m *PingEntry) Reset()                    { *m = PingEntry{} }
func (m *PingEntry) String() string            { return proto.CompactTextString(m) }
func (*PingEntry) ProtoMessage()               {}
func (*PingEntry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PingEntry) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *PingEntry) GetIP() string {
	if m != nil {
		return m.IP
	}
	return ""
}

func (m *PingEntry) GetInterval() uint32 {
	if m != nil {
		return m.Interval
	}
	return 0
}

func (m *PingEntry) GetMetricHostName() string {
	if m != nil {
		return m.MetricHostName
	}
	return ""
}

func (m *PingEntry) GetPingerID() string {
	if m != nil {
		return m.PingerID
	}
	return ""
}

func (m *PingEntry) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *PingEntry) GetIPVersion() uint32 {
	if m != nil {
		return m.IPVersion
	}
	return 0
}

type PingTargetStatus struct {
	IP        string `protobuf:"bytes,1,opt,name=IP" json:"IP,omitempty"`
	Name      string `protobuf:"bytes,2,opt,name=Name" json:"Name,omitempty"`
	Reachable bool   `protobuf:"varint,3,opt,name=Reachable" json:"Reachable,omitempty"`
	Since     uint32 `protobuf:"varint,4,opt,name=Since" json:"Since,omitempty"`
}

func (m *PingTargetStatus) Reset()                    { *m = PingTargetStatus{} }
func (m *PingTargetStatus) String() string            { return proto.CompactTextString(m) }
func (*PingTargetStatus) ProtoMessage()               {}
func (*PingTargetStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *PingTargetStatus) GetIP() string {
	if m != nil {
		return m.IP
	}
	return ""
}

func (m *PingTargetStatus) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *PingTargetStatus) GetReachable() bool {
	if m != nil {
		return m.Reachable
	}
	return false
}

func (m *PingTargetStatus) GetSince() uint32 {
	if m != nil {
		return m.Since
	}
	return 0
}

type PingTargetStatusList struct {
	Status []*PingTargetStatus `protobuf:"bytes,1,rep,name=Status" json:"Status,omitempty"`
}

func (m *PingTargetStatusList) Reset()                    { *m = PingTargetStatusList{} }
func (m *PingTargetStatusList) String() string            { return proto.CompactTextString(m) }
func (*PingTargetStatusList) ProtoMessage()               {}
func (*PingTargetStatusList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *PingTargetStatusList) GetStatus() []*PingTargetStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

type SetPingStatusRequest struct {
	ID      uint64 `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	Success bool   `protobuf:"varint,2,opt,name=Success" json:"Success,omitempty"`
}

func (m *SetPingStatusRequest) Reset()                    { *m = SetPingStatusRequest{} }
func (m *SetPingStatusRequest) String() string            { return proto.CompactTextString(m) }
func (*SetPingStatusRequest) ProtoMessage()               {}
func (*SetPingStatusRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *SetPingStatusRequest) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *SetPingStatusRequest) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

type PingStatus struct {
	PingEntry   *PingEntry `protobuf:"bytes,1,opt,name=PingEntry" json:"PingEntry,omitempty"`
	Currently   bool       `protobuf:"varint,3,opt,name=Currently" json:"Currently,omitempty"`
	Since       uint32     `protobuf:"varint,4,opt,name=Since" json:"Since,omitempty"`
	LastOnline  uint32     `protobuf:"varint,5,opt,name=LastOnline" json:"LastOnline,omitempty"`
	LastOffline uint32     `protobuf:"varint,6,opt,name=LastOffline" json:"LastOffline,omitempty"`
}

func (m *PingStatus) Reset()                    { *m = PingStatus{} }
func (m *PingStatus) String() string            { return proto.CompactTextString(m) }
func (*PingStatus) ProtoMessage()               {}
func (*PingStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *PingStatus) GetPingEntry() *PingEntry {
	if m != nil {
		return m.PingEntry
	}
	return nil
}

func (m *PingStatus) GetCurrently() bool {
	if m != nil {
		return m.Currently
	}
	return false
}

func (m *PingStatus) GetSince() uint32 {
	if m != nil {
		return m.Since
	}
	return 0
}

func (m *PingStatus) GetLastOnline() uint32 {
	if m != nil {
		return m.LastOnline
	}
	return 0
}

func (m *PingStatus) GetLastOffline() uint32 {
	if m != nil {
		return m.LastOffline
	}
	return 0
}

type PingStatusList struct {
	Status []*PingStatus `protobuf:"bytes,1,rep,name=Status" json:"Status,omitempty"`
}

func (m *PingStatusList) Reset()                    { *m = PingStatusList{} }
func (m *PingStatusList) String() string            { return proto.CompactTextString(m) }
func (*PingStatusList) ProtoMessage()               {}
func (*PingStatusList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *PingStatusList) GetStatus() []*PingStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

func init() {
	proto.RegisterType((*PingResult)(nil), "pinger.PingResult")
	proto.RegisterType((*PingRequest)(nil), "pinger.PingRequest")
	proto.RegisterType((*PingListRequest)(nil), "pinger.PingListRequest")
	proto.RegisterType((*PingList)(nil), "pinger.PingList")
	proto.RegisterType((*PingEntry)(nil), "pinger.PingEntry")
	proto.RegisterType((*PingTargetStatus)(nil), "pinger.PingTargetStatus")
	proto.RegisterType((*PingTargetStatusList)(nil), "pinger.PingTargetStatusList")
	proto.RegisterType((*SetPingStatusRequest)(nil), "pinger.SetPingStatusRequest")
	proto.RegisterType((*PingStatus)(nil), "pinger.PingStatus")
	proto.RegisterType((*PingStatusList)(nil), "pinger.PingStatusList")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Pinger service

type PingerClient interface {
	// execute a single ping
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResult, error)
	// get list of current ping targets and their status
	PingStatus(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*PingTargetStatusList, error)
}

type pingerClient struct {
	cc *grpc.ClientConn
}

func NewPingerClient(cc *grpc.ClientConn) PingerClient {
	return &pingerClient{cc}
}

func (c *pingerClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResult, error) {
	out := new(PingResult)
	err := grpc.Invoke(ctx, "/pinger.Pinger/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pingerClient) PingStatus(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*PingTargetStatusList, error) {
	out := new(PingTargetStatusList)
	err := grpc.Invoke(ctx, "/pinger.Pinger/PingStatus", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Pinger service

type PingerServer interface {
	// execute a single ping
	Ping(context.Context, *PingRequest) (*PingResult, error)
	// get list of current ping targets and their status
	PingStatus(context.Context, *common.Void) (*PingTargetStatusList, error)
}

func RegisterPingerServer(s *grpc.Server, srv PingerServer) {
	s.RegisterService(&_Pinger_serviceDesc, srv)
}

func _Pinger_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PingerServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pinger.Pinger/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PingerServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Pinger_PingStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(common.Void)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PingerServer).PingStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pinger.Pinger/PingStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PingerServer).PingStatus(ctx, req.(*common.Void))
	}
	return interceptor(ctx, in, info, handler)
}

var _Pinger_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pinger.Pinger",
	HandlerType: (*PingerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Pinger_Ping_Handler,
		},
		{
			MethodName: "PingStatus",
			Handler:    _Pinger_PingStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/golang.conradwood.net/apis/pinger/pinger.proto",
}

// Client API for PingerList service

type PingerListClient interface {
	// get ping config (list of ip addresses to ping regularly).
	GetPingList(ctx context.Context, in *PingListRequest, opts ...grpc.CallOption) (*PingList, error)
	// pinger reports a status through this RPC
	SetPingStatus(ctx context.Context, in *SetPingStatusRequest, opts ...grpc.CallOption) (*common.Void, error)
	// get a list of all known status'
	GetPingStatus(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*PingStatusList, error)
}

type pingerListClient struct {
	cc *grpc.ClientConn
}

func NewPingerListClient(cc *grpc.ClientConn) PingerListClient {
	return &pingerListClient{cc}
}

func (c *pingerListClient) GetPingList(ctx context.Context, in *PingListRequest, opts ...grpc.CallOption) (*PingList, error) {
	out := new(PingList)
	err := grpc.Invoke(ctx, "/pinger.PingerList/GetPingList", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pingerListClient) SetPingStatus(ctx context.Context, in *SetPingStatusRequest, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/pinger.PingerList/SetPingStatus", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pingerListClient) GetPingStatus(ctx context.Context, in *common.Void, opts ...grpc.CallOption) (*PingStatusList, error) {
	out := new(PingStatusList)
	err := grpc.Invoke(ctx, "/pinger.PingerList/GetPingStatus", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for PingerList service

type PingerListServer interface {
	// get ping config (list of ip addresses to ping regularly).
	GetPingList(context.Context, *PingListRequest) (*PingList, error)
	// pinger reports a status through this RPC
	SetPingStatus(context.Context, *SetPingStatusRequest) (*common.Void, error)
	// get a list of all known status'
	GetPingStatus(context.Context, *common.Void) (*PingStatusList, error)
}

func RegisterPingerListServer(s *grpc.Server, srv PingerListServer) {
	s.RegisterService(&_PingerList_serviceDesc, srv)
}

func _PingerList_GetPingList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PingerListServer).GetPingList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pinger.PingerList/GetPingList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PingerListServer).GetPingList(ctx, req.(*PingListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PingerList_SetPingStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetPingStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PingerListServer).SetPingStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pinger.PingerList/SetPingStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PingerListServer).SetPingStatus(ctx, req.(*SetPingStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PingerList_GetPingStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(common.Void)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PingerListServer).GetPingStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pinger.PingerList/GetPingStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PingerListServer).GetPingStatus(ctx, req.(*common.Void))
	}
	return interceptor(ctx, in, info, handler)
}

var _PingerList_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pinger.PingerList",
	HandlerType: (*PingerListServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPingList",
			Handler:    _PingerList_GetPingList_Handler,
		},
		{
			MethodName: "SetPingStatus",
			Handler:    _PingerList_SetPingStatus_Handler,
		},
		{
			MethodName: "GetPingStatus",
			Handler:    _PingerList_GetPingStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/golang.conradwood.net/apis/pinger/pinger.proto",
}

func init() {
	proto.RegisterFile("protos/golang.conradwood.net/apis/pinger/pinger.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 597 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x54, 0x4f, 0x6f, 0xd3, 0x4e,
	0x10, 0x95, 0xdd, 0xd4, 0x6d, 0x26, 0x75, 0x7e, 0xfd, 0x2d, 0x11, 0x58, 0x56, 0x40, 0x91, 0x0f,
	0x28, 0x02, 0xe1, 0xa0, 0xa0, 0x8a, 0x4a, 0x70, 0x40, 0x50, 0xd4, 0x5a, 0x4a, 0x21, 0x72, 0x50,
	0x0f, 0xdc, 0x1c, 0x67, 0x1b, 0x16, 0x39, 0xbb, 0xc1, 0xbb, 0x01, 0xf5, 0x73, 0x71, 0xe3, 0xc2,
	0x57, 0x43, 0xfb, 0x27, 0xce, 0x3a, 0x09, 0x9c, 0xec, 0x99, 0x79, 0xb3, 0x3b, 0xef, 0xbd, 0xb1,
	0xe1, 0x6c, 0x59, 0x32, 0xc1, 0xf8, 0x60, 0xce, 0x8a, 0x8c, 0xce, 0xe3, 0x9c, 0xd1, 0x32, 0x9b,
	0xfd, 0x60, 0x6c, 0x16, 0x53, 0x2c, 0x06, 0xd9, 0x92, 0xf0, 0xc1, 0x92, 0xd0, 0x39, 0x2e, 0xcd,
	0x23, 0x56, 0x78, 0xe4, 0xe9, 0x28, 0x8c, 0xff, 0xd1, 0x97, 0xb3, 0xc5, 0x82, 0x51, 0xf3, 0xd0,
	0x7d, 0xd1, 0x67, 0x80, 0x31, 0xa1, 0xf3, 0x14, 0xf3, 0x55, 0x21, 0x50, 0x1b, 0xdc, 0x64, 0x1c,
	0x38, 0x3d, 0xa7, 0xdf, 0x4c, 0xdd, 0x64, 0x8c, 0x22, 0x38, 0xb9, 0x26, 0x45, 0x41, 0x38, 0xce,
	0x19, 0x9d, 0xf1, 0xc0, 0xed, 0x39, 0x7d, 0x3f, 0xad, 0xe5, 0x50, 0x00, 0x47, 0x93, 0x55, 0x9e,
	0x63, 0xce, 0x83, 0x83, 0x9e, 0xd3, 0x3f, 0x4e, 0xd7, 0x61, 0xf4, 0x10, 0x5a, 0xfa, 0xec, 0x6f,
	0x2b, 0xcc, 0x77, 0x0e, 0x8f, 0x9e, 0xc1, 0x7f, 0xb2, 0x3c, 0x22, 0x5c, 0xac, 0x21, 0x21, 0x1c,
	0x8f, 0x15, 0x8f, 0xe4, 0xc2, 0x00, 0xab, 0x38, 0x7a, 0xa9, 0x6b, 0x12, 0x8e, 0x9e, 0xc2, 0xd1,
	0x7b, 0x2a, 0x4a, 0x82, 0x79, 0xe0, 0xf4, 0x0e, 0xfa, 0xad, 0xe1, 0xff, 0xb1, 0x51, 0x43, 0x42,
	0x64, 0xe9, 0x2e, 0x5d, 0x23, 0xa2, 0xdf, 0x0e, 0x34, 0xab, 0xb4, 0x9a, 0x42, 0x1f, 0xde, 0x48,
	0xdd, 0xe4, 0xc2, 0x4c, 0xe5, 0x56, 0x94, 0x43, 0x38, 0x4e, 0xa8, 0xc0, 0xe5, 0xf7, 0xac, 0x50,
	0x7c, 0xfc, 0xb4, 0x8a, 0xd1, 0x63, 0x68, 0x5f, 0x63, 0x51, 0x92, 0xfc, 0x8a, 0x71, 0xf1, 0x21,
	0x5b, 0xe0, 0xa0, 0xa1, 0xfa, 0xb6, 0xb2, 0x35, 0x1a, 0x87, 0x75, 0x1a, 0xa8, 0x03, 0x87, 0xa3,
	0x6c, 0x8a, 0x8b, 0xc0, 0x53, 0x05, 0x1d, 0xa0, 0x2e, 0x34, 0x93, 0xf1, 0x0d, 0x2e, 0x39, 0x61,
	0x34, 0x38, 0x52, 0xd7, 0x6e, 0x12, 0xd1, 0x57, 0x38, 0x95, 0xfd, 0x9f, 0xb2, 0x72, 0x8e, 0xc5,
	0x44, 0x64, 0x62, 0xc5, 0x77, 0xac, 0x42, 0xd0, 0x50, 0x13, 0x69, 0x26, 0xea, 0x5d, 0x9e, 0x9a,
	0xe2, 0x2c, 0xff, 0x92, 0x4d, 0x0b, 0x6c, 0xcc, 0xd9, 0x24, 0xe4, 0x24, 0x13, 0x42, 0x73, 0x4d,
	0xc2, 0x4f, 0x75, 0x10, 0x5d, 0x41, 0x67, 0xfb, 0x2e, 0x25, 0xf9, 0x73, 0xf0, 0x74, 0x64, 0x14,
	0x0f, 0x6c, 0xc5, 0x6d, 0x74, 0x6a, 0x70, 0xd1, 0x1b, 0xe8, 0x4c, 0xb0, 0x90, 0x65, 0x53, 0xb0,
	0xf6, 0xc0, 0x76, 0xc0, 0x5a, 0x20, 0xb7, 0xbe, 0x40, 0x3f, 0x1d, 0xbd, 0x9d, 0x86, 0xf2, 0xc0,
	0xf2, 0x51, 0xf5, 0xef, 0xf5, 0xdd, 0xf2, 0xba, 0x0b, 0xcd, 0x77, 0xab, 0xb2, 0xc4, 0x54, 0x14,
	0x77, 0x6b, 0xfe, 0x55, 0x62, 0x3f, 0x7f, 0xf4, 0x08, 0x60, 0x94, 0x71, 0xf1, 0x91, 0x16, 0x84,
	0x62, 0xe5, 0x9e, 0x9f, 0x5a, 0x19, 0xd4, 0x83, 0x96, 0x8a, 0x6e, 0x6f, 0x15, 0xc0, 0x53, 0x00,
	0x3b, 0x15, 0xbd, 0x86, 0xf6, 0x66, 0x68, 0xa5, 0xdd, 0x93, 0x2d, 0xed, 0x90, 0x3d, 0x75, 0x5d,
	0xb5, 0x21, 0x07, 0x4f, 0xef, 0x0a, 0x1a, 0x40, 0x43, 0xbe, 0xa1, 0x7b, 0x36, 0xda, 0x88, 0x18,
	0xa2, 0x7a, 0x52, 0x7d, 0xbd, 0xe7, 0x35, 0xb5, 0x4e, 0x62, 0xf3, 0xa1, 0xdf, 0x30, 0x32, 0x0b,
	0xbb, 0x7f, 0xb3, 0x4b, 0x0e, 0x38, 0xfc, 0x65, 0x84, 0xc6, 0xa5, 0x9a, 0xf7, 0x1c, 0x5a, 0x97,
	0xda, 0x39, 0x15, 0x3e, 0xb0, 0x7b, 0xad, 0xcf, 0x35, 0x3c, 0xdd, 0x2e, 0xa0, 0x57, 0xe0, 0xd7,
	0x3c, 0x47, 0xd5, 0xbd, 0xfb, 0x56, 0x21, 0xac, 0xcd, 0x88, 0xce, 0xc0, 0xbf, 0xac, 0x35, 0xd7,
	0x29, 0xdc, 0xdf, 0x55, 0x4d, 0xde, 0xf9, 0xb6, 0x0b, 0x21, 0xc5, 0xc2, 0xfe, 0xe3, 0xc9, 0xbf,
	0x9d, 0x01, 0x4f, 0x3d, 0xf5, 0x9f, 0x7b, 0xf1, 0x27, 0x00, 0x00, 0xff, 0xff, 0x60, 0x50, 0x63,
	0xfc, 0x58, 0x05, 0x00, 0x00,
}

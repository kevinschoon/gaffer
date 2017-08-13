// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/mesanine/gaffer/plugin/rpc-server/rpc.proto

/*
Package server is a generated protocol buffer package.

It is generated from these files:
	github.com/mesanine/gaffer/plugin/rpc-server/rpc.proto

It has these top-level messages:
	StatusRequest
	StatusResponse
	RestartRequest
	RestartResponse
*/
package server

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import host "github.com/mesanine/gaffer/host"
import service "github.com/mesanine/gaffer/service"

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

type StatusRequest struct {
	Host *host.Host `protobuf:"bytes,1,opt,name=host" json:"host,omitempty"`
}

func (m *StatusRequest) Reset()                    { *m = StatusRequest{} }
func (m *StatusRequest) String() string            { return proto.CompactTextString(m) }
func (*StatusRequest) ProtoMessage()               {}
func (*StatusRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *StatusRequest) GetHost() *host.Host {
	if m != nil {
		return m.Host
	}
	return nil
}

type StatusResponse struct {
	Services []*service.Service `protobuf:"bytes,2,rep,name=services" json:"services,omitempty"`
}

func (m *StatusResponse) Reset()                    { *m = StatusResponse{} }
func (m *StatusResponse) String() string            { return proto.CompactTextString(m) }
func (*StatusResponse) ProtoMessage()               {}
func (*StatusResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *StatusResponse) GetServices() []*service.Service {
	if m != nil {
		return m.Services
	}
	return nil
}

type RestartRequest struct {
	Id   string     `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Host *host.Host `protobuf:"bytes,2,opt,name=host" json:"host,omitempty"`
}

func (m *RestartRequest) Reset()                    { *m = RestartRequest{} }
func (m *RestartRequest) String() string            { return proto.CompactTextString(m) }
func (*RestartRequest) ProtoMessage()               {}
func (*RestartRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *RestartRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *RestartRequest) GetHost() *host.Host {
	if m != nil {
		return m.Host
	}
	return nil
}

type RestartResponse struct {
}

func (m *RestartResponse) Reset()                    { *m = RestartResponse{} }
func (m *RestartResponse) String() string            { return proto.CompactTextString(m) }
func (*RestartResponse) ProtoMessage()               {}
func (*RestartResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func init() {
	proto.RegisterType((*StatusRequest)(nil), "server.StatusRequest")
	proto.RegisterType((*StatusResponse)(nil), "server.StatusResponse")
	proto.RegisterType((*RestartRequest)(nil), "server.RestartRequest")
	proto.RegisterType((*RestartResponse)(nil), "server.RestartResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for RPC service

type RPCClient interface {
	// Status returns the status of a given container service
	Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	// Restart restarts a container service
	Restart(ctx context.Context, in *RestartRequest, opts ...grpc.CallOption) (*RestartResponse, error)
}

type rPCClient struct {
	cc *grpc.ClientConn
}

func NewRPCClient(cc *grpc.ClientConn) RPCClient {
	return &rPCClient{cc}
}

func (c *rPCClient) Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := grpc.Invoke(ctx, "/server.RPC/Status", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rPCClient) Restart(ctx context.Context, in *RestartRequest, opts ...grpc.CallOption) (*RestartResponse, error) {
	out := new(RestartResponse)
	err := grpc.Invoke(ctx, "/server.RPC/Restart", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for RPC service

type RPCServer interface {
	// Status returns the status of a given container service
	Status(context.Context, *StatusRequest) (*StatusResponse, error)
	// Restart restarts a container service
	Restart(context.Context, *RestartRequest) (*RestartResponse, error)
}

func RegisterRPCServer(s *grpc.Server, srv RPCServer) {
	s.RegisterService(&_RPC_serviceDesc, srv)
}

func _RPC_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/server.RPC/Status",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCServer).Status(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RPC_Restart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RestartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCServer).Restart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/server.RPC/Restart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCServer).Restart(ctx, req.(*RestartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _RPC_serviceDesc = grpc.ServiceDesc{
	ServiceName: "server.RPC",
	HandlerType: (*RPCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Status",
			Handler:    _RPC_Status_Handler,
		},
		{
			MethodName: "Restart",
			Handler:    _RPC_Restart_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/mesanine/gaffer/plugin/rpc-server/rpc.proto",
}

func init() {
	proto.RegisterFile("github.com/mesanine/gaffer/plugin/rpc-server/rpc.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 270 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x90, 0x51, 0x4b, 0xc3, 0x30,
	0x14, 0x85, 0x5d, 0x27, 0x55, 0xef, 0xb0, 0x6a, 0x40, 0x1d, 0x7d, 0x90, 0xd1, 0xa7, 0x21, 0x9a,
	0x48, 0x05, 0x41, 0x10, 0x11, 0x7c, 0xf1, 0x51, 0xb2, 0x5f, 0xd0, 0x75, 0x77, 0x5d, 0xc0, 0x35,
	0x35, 0x37, 0xf5, 0xd1, 0xdf, 0x2e, 0x4d, 0x96, 0xc2, 0x14, 0xf7, 0xd2, 0x5b, 0x0e, 0xf7, 0x7c,
	0xe7, 0xe4, 0xc2, 0x43, 0xa5, 0xec, 0xaa, 0x9d, 0xf3, 0x52, 0xaf, 0xc5, 0x1a, 0xa9, 0xa8, 0x55,
	0x8d, 0xa2, 0x2a, 0x96, 0x4b, 0x34, 0xa2, 0xf9, 0x68, 0x2b, 0x55, 0x0b, 0xd3, 0x94, 0xb7, 0x84,
	0xe6, 0x0b, 0x4d, 0xf7, 0xcb, 0x1b, 0xa3, 0xad, 0x66, 0xb1, 0x57, 0xd2, 0xeb, 0x1d, 0xfe, 0x95,
	0x26, 0xeb, 0x3e, 0xde, 0x93, 0xde, 0xed, 0xd8, 0xed, 0x70, 0xaa, 0xc4, 0x30, 0xbd, 0x23, 0x13,
	0x70, 0x3c, 0xb3, 0x85, 0x6d, 0x49, 0xe2, 0x67, 0x8b, 0x64, 0xd9, 0x15, 0xec, 0x77, 0xc0, 0xf1,
	0x60, 0x32, 0x98, 0x8e, 0x72, 0xe0, 0x8e, 0xfe, 0xa6, 0xc9, 0x4a, 0xa7, 0x67, 0xcf, 0x90, 0x04,
	0x03, 0x35, 0xba, 0x26, 0x64, 0x37, 0x70, 0xb8, 0x61, 0xd2, 0x38, 0x9a, 0x0c, 0xa7, 0xa3, 0xfc,
	0x94, 0x87, 0x90, 0x99, 0x9f, 0xb2, 0xdf, 0xc8, 0x5e, 0x20, 0x91, 0x48, 0xb6, 0x30, 0x36, 0x24,
	0x26, 0x10, 0xa9, 0x85, 0xcb, 0x3b, 0x92, 0x91, 0x5a, 0xf4, 0x0d, 0xa2, 0x7f, 0x1a, 0x9c, 0xc1,
	0x49, 0x4f, 0xf0, 0x15, 0xf2, 0x6f, 0x18, 0xca, 0xf7, 0x57, 0xf6, 0x08, 0xb1, 0xef, 0xc6, 0xce,
	0xb9, 0xbf, 0x1e, 0xdf, 0x7a, 0x5c, 0x7a, 0xf1, 0x5b, 0xf6, 0xfe, 0x6c, 0x8f, 0x3d, 0xc1, 0xc1,
	0x06, 0xca, 0xfa, 0xa5, 0xed, 0x9e, 0xe9, 0xe5, 0x1f, 0x3d, 0xb8, 0xe7, 0xb1, 0x3b, 0xe6, 0xfd,
	0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x64, 0x7f, 0xc0, 0xe1, 0xec, 0x01, 0x00, 0x00,
}

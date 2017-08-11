// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/mesanine/gaffer/plugin/supervisor/supervisor.proto

/*
Package supervisor is a generated protocol buffer package.

It is generated from these files:
	github.com/mesanine/gaffer/plugin/supervisor/supervisor.proto

It has these top-level messages:
	StatusRequest
	StatusResponse
	RestartRequest
	RestartResponse
*/
package supervisor

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/any"
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
}

func (m *StatusRequest) Reset()                    { *m = StatusRequest{} }
func (m *StatusRequest) String() string            { return proto.CompactTextString(m) }
func (*StatusRequest) ProtoMessage()               {}
func (*StatusRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type StatusResponse struct {
	Host     *host.Host                      `protobuf:"bytes,1,opt,name=host" json:"host,omitempty"`
	Services map[string]*service.Service     `protobuf:"bytes,2,rep,name=services" json:"services,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Stats    map[string]*google_protobuf.Any `protobuf:"bytes,3,rep,name=stats" json:"stats,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *StatusResponse) Reset()                    { *m = StatusResponse{} }
func (m *StatusResponse) String() string            { return proto.CompactTextString(m) }
func (*StatusResponse) ProtoMessage()               {}
func (*StatusResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *StatusResponse) GetHost() *host.Host {
	if m != nil {
		return m.Host
	}
	return nil
}

func (m *StatusResponse) GetServices() map[string]*service.Service {
	if m != nil {
		return m.Services
	}
	return nil
}

func (m *StatusResponse) GetStats() map[string]*google_protobuf.Any {
	if m != nil {
		return m.Stats
	}
	return nil
}

type RestartRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
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

type RestartResponse struct {
}

func (m *RestartResponse) Reset()                    { *m = RestartResponse{} }
func (m *RestartResponse) String() string            { return proto.CompactTextString(m) }
func (*RestartResponse) ProtoMessage()               {}
func (*RestartResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func init() {
	proto.RegisterType((*StatusRequest)(nil), "supervisor.StatusRequest")
	proto.RegisterType((*StatusResponse)(nil), "supervisor.StatusResponse")
	proto.RegisterType((*RestartRequest)(nil), "supervisor.RestartRequest")
	proto.RegisterType((*RestartResponse)(nil), "supervisor.RestartResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Supervisor service

type SupervisorClient interface {
	// Status returns the status of a given container service
	Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	// Restart restarts a container service
	Restart(ctx context.Context, in *RestartRequest, opts ...grpc.CallOption) (*RestartResponse, error)
}

type supervisorClient struct {
	cc *grpc.ClientConn
}

func NewSupervisorClient(cc *grpc.ClientConn) SupervisorClient {
	return &supervisorClient{cc}
}

func (c *supervisorClient) Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := grpc.Invoke(ctx, "/supervisor.Supervisor/Status", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *supervisorClient) Restart(ctx context.Context, in *RestartRequest, opts ...grpc.CallOption) (*RestartResponse, error) {
	out := new(RestartResponse)
	err := grpc.Invoke(ctx, "/supervisor.Supervisor/Restart", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Supervisor service

type SupervisorServer interface {
	// Status returns the status of a given container service
	Status(context.Context, *StatusRequest) (*StatusResponse, error)
	// Restart restarts a container service
	Restart(context.Context, *RestartRequest) (*RestartResponse, error)
}

func RegisterSupervisorServer(s *grpc.Server, srv SupervisorServer) {
	s.RegisterService(&_Supervisor_serviceDesc, srv)
}

func _Supervisor_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.Supervisor/Status",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).Status(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Supervisor_Restart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RestartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).Restart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.Supervisor/Restart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).Restart(ctx, req.(*RestartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Supervisor_serviceDesc = grpc.ServiceDesc{
	ServiceName: "supervisor.Supervisor",
	HandlerType: (*SupervisorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Status",
			Handler:    _Supervisor_Status_Handler,
		},
		{
			MethodName: "Restart",
			Handler:    _Supervisor_Restart_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/mesanine/gaffer/plugin/supervisor/supervisor.proto",
}

func init() {
	proto.RegisterFile("github.com/mesanine/gaffer/plugin/supervisor/supervisor.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 371 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0xcf, 0x4e, 0xf2, 0x40,
	0x14, 0xc5, 0xbf, 0x96, 0x0f, 0xd4, 0x4b, 0xf8, 0xe3, 0xc4, 0x45, 0xa9, 0x89, 0x21, 0x4d, 0x34,
	0x84, 0xc5, 0xd4, 0xe0, 0xc6, 0x68, 0x5c, 0x90, 0x60, 0xe2, 0x46, 0x17, 0xe5, 0x09, 0x0a, 0x0c,
	0xa5, 0x11, 0x3a, 0xb5, 0x77, 0x86, 0xa4, 0x0f, 0xe2, 0x8b, 0xf9, 0x44, 0xa6, 0x33, 0x53, 0xa0,
	0x09, 0xb2, 0x69, 0x27, 0xf7, 0x9e, 0xfb, 0x3b, 0x73, 0xcf, 0xc0, 0x4b, 0x14, 0x8b, 0x95, 0x9c,
	0xd1, 0x39, 0xdf, 0xf8, 0x1b, 0x86, 0x61, 0x12, 0x27, 0xcc, 0x8f, 0xc2, 0xe5, 0x92, 0x65, 0x7e,
	0xba, 0x96, 0x51, 0x9c, 0xf8, 0x28, 0x53, 0x96, 0x6d, 0x63, 0xe4, 0xd9, 0xc1, 0x91, 0xa6, 0x19,
	0x17, 0x9c, 0xc0, 0xbe, 0xe2, 0xf6, 0x22, 0xce, 0xa3, 0x35, 0xf3, 0x55, 0x67, 0x26, 0x97, 0x7e,
	0x98, 0xe4, 0x5a, 0xe6, 0x0e, 0x4f, 0xb8, 0xac, 0x38, 0x0a, 0xf5, 0x31, 0xda, 0xfb, 0x13, 0x5a,
	0x2c, 0xbc, 0xe6, 0xac, 0xfc, 0xeb, 0x09, 0xaf, 0x03, 0xad, 0xa9, 0x08, 0x85, 0xc4, 0x80, 0x7d,
	0x49, 0x86, 0xc2, 0xfb, 0xb1, 0xa1, 0x5d, 0x56, 0x30, 0xe5, 0x09, 0x32, 0x72, 0x03, 0xff, 0x0b,
	0x0f, 0xc7, 0xea, 0x5b, 0x83, 0xe6, 0x08, 0xa8, 0x32, 0x7c, 0xe3, 0x28, 0x02, 0x55, 0x27, 0x13,
	0x38, 0x37, 0x50, 0x74, 0xec, 0x7e, 0x6d, 0xd0, 0x1c, 0x0d, 0xe8, 0xc1, 0xb6, 0x55, 0x1a, 0x9d,
	0x1a, 0xe9, 0x6b, 0x22, 0xb2, 0x3c, 0xd8, 0x4d, 0x92, 0x67, 0xa8, 0xa3, 0x08, 0x05, 0x3a, 0x35,
	0x85, 0xb8, 0x3d, 0x85, 0x28, 0x74, 0x7a, 0x5e, 0xcf, 0xb8, 0xef, 0xd0, 0xaa, 0x70, 0x49, 0x17,
	0x6a, 0x9f, 0x2c, 0x57, 0x57, 0xbe, 0x08, 0x8a, 0x23, 0xb9, 0x83, 0xfa, 0x36, 0x5c, 0x4b, 0xe6,
	0xd8, 0x6a, 0x8d, 0x2e, 0x2d, 0x83, 0x30, 0x83, 0x81, 0x6e, 0x3f, 0xd9, 0x8f, 0x96, 0xfb, 0x01,
	0xb0, 0xf7, 0x38, 0xc2, 0x1a, 0x56, 0x59, 0x57, 0x54, 0x3f, 0x1f, 0x2d, 0x9f, 0x8f, 0x8e, 0x93,
	0xfc, 0x80, 0xe7, 0xf5, 0xa1, 0x1d, 0x30, 0x14, 0x61, 0x26, 0x4c, 0xcc, 0xa4, 0x0d, 0x76, 0xbc,
	0x30, 0x48, 0x3b, 0x5e, 0x78, 0x97, 0xd0, 0xd9, 0x29, 0xf4, 0x96, 0xa3, 0x6f, 0x0b, 0x60, 0xba,
	0xcb, 0x80, 0x8c, 0xa1, 0xa1, 0x63, 0x20, 0xbd, 0x63, 0xd1, 0x28, 0xac, 0xeb, 0xfe, 0x9d, 0x9a,
	0xf7, 0x8f, 0x4c, 0xe0, 0xcc, 0x98, 0x90, 0x8a, 0xb0, 0x7a, 0x37, 0xf7, 0xfa, 0x68, 0xaf, 0xa4,
	0xcc, 0x1a, 0x6a, 0xcb, 0x87, 0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x57, 0xd8, 0xb9, 0xa6, 0xff,
	0x02, 0x00, 0x00,
}
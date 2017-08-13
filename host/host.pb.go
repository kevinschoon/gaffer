// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/mesanine/gaffer/host/host.proto

/*
Package host is a generated protocol buffer package.

It is generated from these files:
	github.com/mesanine/gaffer/host/host.proto

It has these top-level messages:
	Host
*/
package host

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Host is a unqiue server within a Gaffer cluster.
type Host struct {
	// Server hostname
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// Server IP address
	Address string `protobuf:"bytes,2,opt,name=address" json:"address,omitempty"`
	// MAC Address
	Mac string `protobuf:"bytes,3,opt,name=mac" json:"mac,omitempty"`
}

func (m *Host) Reset()                    { *m = Host{} }
func (m *Host) String() string            { return proto.CompactTextString(m) }
func (*Host) ProtoMessage()               {}
func (*Host) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Host) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Host) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Host) GetMac() string {
	if m != nil {
		return m.Mac
	}
	return ""
}

func init() {
	proto.RegisterType((*Host)(nil), "host.Host")
}

func init() { proto.RegisterFile("github.com/mesanine/gaffer/host/host.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 128 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xd2, 0x4a, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0xcf, 0x4d, 0x2d, 0x4e, 0xcc, 0xcb, 0xcc, 0x4b, 0xd5,
	0x4f, 0x4f, 0x4c, 0x4b, 0x4b, 0x2d, 0xd2, 0xcf, 0xc8, 0x2f, 0x2e, 0x01, 0x13, 0x7a, 0x05, 0x45,
	0xf9, 0x25, 0xf9, 0x42, 0x2c, 0x20, 0xb6, 0x92, 0x1b, 0x17, 0x8b, 0x47, 0x7e, 0x71, 0x89, 0x90,
	0x10, 0x17, 0x4b, 0x5e, 0x62, 0x6e, 0xaa, 0x04, 0xa3, 0x02, 0xa3, 0x06, 0x67, 0x10, 0x98, 0x2d,
	0x24, 0xc1, 0xc5, 0x9e, 0x98, 0x92, 0x52, 0x94, 0x5a, 0x5c, 0x2c, 0xc1, 0x04, 0x16, 0x86, 0x71,
	0x85, 0x04, 0xb8, 0x98, 0x73, 0x13, 0x93, 0x25, 0x98, 0xc1, 0xa2, 0x20, 0x66, 0x12, 0x1b, 0xd8,
	0x50, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0xcc, 0x70, 0x0a, 0xfd, 0x82, 0x00, 0x00, 0x00,
}

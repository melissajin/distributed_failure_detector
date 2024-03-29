// Code generated by protoc-gen-go. DO NOT EDIT.
// source: heartbeat/heartbeat.proto

/*
Package heartbeat is a generated protocol buffer package.

It is generated from these files:
	heartbeat/heartbeat.proto

It has these top-level messages:
	Machine
	Heartbeat
*/
package heartbeat

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

type Machine struct {
	Id        *Machine_Id `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	HbCounter int32       `protobuf:"varint,2,opt,name=hbCounter" json:"hbCounter,omitempty"`
	Status    int32       `protobuf:"varint,3,opt,name=status" json:"status,omitempty"`
}

func (m *Machine) Reset()                    { *m = Machine{} }
func (m *Machine) String() string            { return proto.CompactTextString(m) }
func (*Machine) ProtoMessage()               {}
func (*Machine) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Machine) GetId() *Machine_Id {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *Machine) GetHbCounter() int32 {
	if m != nil {
		return m.HbCounter
	}
	return 0
}

func (m *Machine) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

type Machine_Id struct {
	Id        int32  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Timestamp string `protobuf:"bytes,2,opt,name=timestamp" json:"timestamp,omitempty"`
}

func (m *Machine_Id) Reset()                    { *m = Machine_Id{} }
func (m *Machine_Id) String() string            { return proto.CompactTextString(m) }
func (*Machine_Id) ProtoMessage()               {}
func (*Machine_Id) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

func (m *Machine_Id) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Machine_Id) GetTimestamp() string {
	if m != nil {
		return m.Timestamp
	}
	return ""
}

type Heartbeat struct {
	Id      int32      `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Machine []*Machine `protobuf:"bytes,2,rep,name=machine" json:"machine,omitempty"`
}

func (m *Heartbeat) Reset()                    { *m = Heartbeat{} }
func (m *Heartbeat) String() string            { return proto.CompactTextString(m) }
func (*Heartbeat) ProtoMessage()               {}
func (*Heartbeat) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Heartbeat) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Heartbeat) GetMachine() []*Machine {
	if m != nil {
		return m.Machine
	}
	return nil
}

func init() {
	proto.RegisterType((*Machine)(nil), "heartbeat.Machine")
	proto.RegisterType((*Machine_Id)(nil), "heartbeat.Machine.Id")
	proto.RegisterType((*Heartbeat)(nil), "heartbeat.Heartbeat")
}

func init() { proto.RegisterFile("heartbeat/heartbeat.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 181 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x92, 0xcc, 0x48, 0x4d, 0x2c,
	0x2a, 0x49, 0x4a, 0x4d, 0x2c, 0xd1, 0x87, 0xb3, 0xf4, 0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0x38,
	0xe1, 0x02, 0x4a, 0xb3, 0x18, 0xb9, 0xd8, 0x7d, 0x13, 0x93, 0x33, 0x32, 0xf3, 0x52, 0x85, 0x54,
	0xb9, 0x98, 0x32, 0x53, 0x24, 0x18, 0x15, 0x18, 0x35, 0xb8, 0x8d, 0x44, 0xf5, 0x10, 0x9a, 0xa0,
	0xf2, 0x7a, 0x9e, 0x29, 0x41, 0x40, 0x05, 0x42, 0x32, 0x5c, 0x9c, 0x19, 0x49, 0xce, 0xf9, 0xa5,
	0x79, 0x25, 0xa9, 0x45, 0x12, 0x4c, 0x40, 0xd5, 0xac, 0x41, 0x08, 0x01, 0x21, 0x31, 0x2e, 0xb6,
	0xe2, 0x92, 0xc4, 0x92, 0xd2, 0x62, 0x09, 0x66, 0xb0, 0x14, 0x94, 0x27, 0x65, 0xc4, 0xc5, 0xe4,
	0x99, 0x22, 0xc4, 0x07, 0xb7, 0x82, 0x15, 0x66, 0x56, 0x49, 0x66, 0x6e, 0x2a, 0x50, 0x4d, 0x6e,
	0x01, 0xd8, 0x2c, 0xce, 0x20, 0x84, 0x80, 0x92, 0x27, 0x17, 0xa7, 0x07, 0xcc, 0x15, 0x18, 0x5a,
	0x75, 0xb8, 0xd8, 0x73, 0x21, 0x0e, 0x03, 0x6a, 0x64, 0x06, 0x3a, 0x59, 0x08, 0xd3, 0xc9, 0x41,
	0x30, 0x25, 0x49, 0x6c, 0x60, 0x9f, 0x1b, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0xa7, 0x9b, 0x44,
	0x22, 0x16, 0x01, 0x00, 0x00,
}

// Code generated by protoc-gen-go.
// source: paranoidnetwork/paranoid.proto
// DO NOT EDIT!

/*
Package paranoid is a generated protocol buffer package.

It is generated from these files:
	paranoidnetwork/paranoid.proto

It has these top-level messages:
	EmptyMessage
	PingRequest
	CreatRequest
	WriteRequest
	WriteResponse
	LinkRequest
	UnlinkRequest
	RenameRequest
	TruncateRequest
	UtimesRequest
	ChmodRequest
*/
package paranoid

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type EmptyMessage struct {
}

func (m *EmptyMessage) Reset()                    { *m = EmptyMessage{} }
func (m *EmptyMessage) String() string            { return proto.CompactTextString(m) }
func (*EmptyMessage) ProtoMessage()               {}
func (*EmptyMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type PingRequest struct {
	Ip         string `protobuf:"bytes,1,opt,name=ip" json:"ip,omitempty"`
	Port       string `protobuf:"bytes,2,opt,name=port" json:"port,omitempty"`
	CommonName string `protobuf:"bytes,3,opt,name=common_name" json:"common_name,omitempty"`
}

func (m *PingRequest) Reset()                    { *m = PingRequest{} }
func (m *PingRequest) String() string            { return proto.CompactTextString(m) }
func (*PingRequest) ProtoMessage()               {}
func (*PingRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type CreatRequest struct {
	Path        string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
	Permissions uint32 `protobuf:"varint,2,opt,name=permissions" json:"permissions,omitempty"`
}

func (m *CreatRequest) Reset()                    { *m = CreatRequest{} }
func (m *CreatRequest) String() string            { return proto.CompactTextString(m) }
func (*CreatRequest) ProtoMessage()               {}
func (*CreatRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type WriteRequest struct {
	Path   string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
	Data   []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Offset uint64 `protobuf:"varint,3,opt,name=offset" json:"offset,omitempty"`
	Length uint64 `protobuf:"varint,4,opt,name=length" json:"length,omitempty"`
}

func (m *WriteRequest) Reset()                    { *m = WriteRequest{} }
func (m *WriteRequest) String() string            { return proto.CompactTextString(m) }
func (*WriteRequest) ProtoMessage()               {}
func (*WriteRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type WriteResponse struct {
	BytesWritten uint64 `protobuf:"varint,1,opt,name=bytes_written" json:"bytes_written,omitempty"`
}

func (m *WriteResponse) Reset()                    { *m = WriteResponse{} }
func (m *WriteResponse) String() string            { return proto.CompactTextString(m) }
func (*WriteResponse) ProtoMessage()               {}
func (*WriteResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type LinkRequest struct {
	OldPath string `protobuf:"bytes,1,opt,name=old_path" json:"old_path,omitempty"`
	NewPath string `protobuf:"bytes,2,opt,name=new_path" json:"new_path,omitempty"`
}

func (m *LinkRequest) Reset()                    { *m = LinkRequest{} }
func (m *LinkRequest) String() string            { return proto.CompactTextString(m) }
func (*LinkRequest) ProtoMessage()               {}
func (*LinkRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type UnlinkRequest struct {
	Path string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
}

func (m *UnlinkRequest) Reset()                    { *m = UnlinkRequest{} }
func (m *UnlinkRequest) String() string            { return proto.CompactTextString(m) }
func (*UnlinkRequest) ProtoMessage()               {}
func (*UnlinkRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type RenameRequest struct {
	OldPath string `protobuf:"bytes,1,opt,name=old_path" json:"old_path,omitempty"`
	NewPath string `protobuf:"bytes,2,opt,name=new_path" json:"new_path,omitempty"`
}

func (m *RenameRequest) Reset()                    { *m = RenameRequest{} }
func (m *RenameRequest) String() string            { return proto.CompactTextString(m) }
func (*RenameRequest) ProtoMessage()               {}
func (*RenameRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type TruncateRequest struct {
	Path   string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
	Length uint64 `protobuf:"varint,2,opt,name=length" json:"length,omitempty"`
}

func (m *TruncateRequest) Reset()                    { *m = TruncateRequest{} }
func (m *TruncateRequest) String() string            { return proto.CompactTextString(m) }
func (*TruncateRequest) ProtoMessage()               {}
func (*TruncateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

// TODO(terry): Update this message to use microseconds by the end of
// sprint 2.
type UtimesRequest struct {
	Path               string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
	AccessSeconds      uint64 `protobuf:"varint,2,opt,name=access_seconds" json:"access_seconds,omitempty"`
	AccessMicroseconds uint64 `protobuf:"varint,3,opt,name=access_microseconds" json:"access_microseconds,omitempty"`
	ModifySeconds      uint64 `protobuf:"varint,4,opt,name=modify_seconds" json:"modify_seconds,omitempty"`
	ModifyMicroseconds uint64 `protobuf:"varint,5,opt,name=modify_microseconds" json:"modify_microseconds,omitempty"`
}

func (m *UtimesRequest) Reset()                    { *m = UtimesRequest{} }
func (m *UtimesRequest) String() string            { return proto.CompactTextString(m) }
func (*UtimesRequest) ProtoMessage()               {}
func (*UtimesRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

type ChmodRequest struct {
	Path string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
	Mode uint32 `protobuf:"varint,2,opt,name=mode" json:"mode,omitempty"`
}

func (m *ChmodRequest) Reset()                    { *m = ChmodRequest{} }
func (m *ChmodRequest) String() string            { return proto.CompactTextString(m) }
func (*ChmodRequest) ProtoMessage()               {}
func (*ChmodRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func init() {
	proto.RegisterType((*EmptyMessage)(nil), "paranoid.EmptyMessage")
	proto.RegisterType((*PingRequest)(nil), "paranoid.PingRequest")
	proto.RegisterType((*CreatRequest)(nil), "paranoid.CreatRequest")
	proto.RegisterType((*WriteRequest)(nil), "paranoid.WriteRequest")
	proto.RegisterType((*WriteResponse)(nil), "paranoid.WriteResponse")
	proto.RegisterType((*LinkRequest)(nil), "paranoid.LinkRequest")
	proto.RegisterType((*UnlinkRequest)(nil), "paranoid.UnlinkRequest")
	proto.RegisterType((*RenameRequest)(nil), "paranoid.RenameRequest")
	proto.RegisterType((*TruncateRequest)(nil), "paranoid.TruncateRequest")
	proto.RegisterType((*UtimesRequest)(nil), "paranoid.UtimesRequest")
	proto.RegisterType((*ChmodRequest)(nil), "paranoid.ChmodRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Client API for ParanoidNetwork service

type ParanoidNetworkClient interface {
	// Used for health checking and discovery. Sends the IP and port of the
	// PFSD instance running on the client.
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*EmptyMessage, error)
	// Filesystem calls
	Creat(ctx context.Context, in *CreatRequest, opts ...grpc.CallOption) (*EmptyMessage, error)
	Write(ctx context.Context, in *WriteRequest, opts ...grpc.CallOption) (*WriteResponse, error)
	Link(ctx context.Context, in *LinkRequest, opts ...grpc.CallOption) (*EmptyMessage, error)
	Unlink(ctx context.Context, in *UnlinkRequest, opts ...grpc.CallOption) (*EmptyMessage, error)
	Rename(ctx context.Context, in *RenameRequest, opts ...grpc.CallOption) (*EmptyMessage, error)
	Truncate(ctx context.Context, in *TruncateRequest, opts ...grpc.CallOption) (*EmptyMessage, error)
	Utimes(ctx context.Context, in *UtimesRequest, opts ...grpc.CallOption) (*EmptyMessage, error)
	Chmod(ctx context.Context, in *ChmodRequest, opts ...grpc.CallOption) (*EmptyMessage, error)
}

type paranoidNetworkClient struct {
	cc *grpc.ClientConn
}

func NewParanoidNetworkClient(cc *grpc.ClientConn) ParanoidNetworkClient {
	return &paranoidNetworkClient{cc}
}

func (c *paranoidNetworkClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*EmptyMessage, error) {
	out := new(EmptyMessage)
	err := grpc.Invoke(ctx, "/paranoid.ParanoidNetwork/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paranoidNetworkClient) Creat(ctx context.Context, in *CreatRequest, opts ...grpc.CallOption) (*EmptyMessage, error) {
	out := new(EmptyMessage)
	err := grpc.Invoke(ctx, "/paranoid.ParanoidNetwork/Creat", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paranoidNetworkClient) Write(ctx context.Context, in *WriteRequest, opts ...grpc.CallOption) (*WriteResponse, error) {
	out := new(WriteResponse)
	err := grpc.Invoke(ctx, "/paranoid.ParanoidNetwork/Write", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paranoidNetworkClient) Link(ctx context.Context, in *LinkRequest, opts ...grpc.CallOption) (*EmptyMessage, error) {
	out := new(EmptyMessage)
	err := grpc.Invoke(ctx, "/paranoid.ParanoidNetwork/Link", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paranoidNetworkClient) Unlink(ctx context.Context, in *UnlinkRequest, opts ...grpc.CallOption) (*EmptyMessage, error) {
	out := new(EmptyMessage)
	err := grpc.Invoke(ctx, "/paranoid.ParanoidNetwork/Unlink", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paranoidNetworkClient) Rename(ctx context.Context, in *RenameRequest, opts ...grpc.CallOption) (*EmptyMessage, error) {
	out := new(EmptyMessage)
	err := grpc.Invoke(ctx, "/paranoid.ParanoidNetwork/Rename", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paranoidNetworkClient) Truncate(ctx context.Context, in *TruncateRequest, opts ...grpc.CallOption) (*EmptyMessage, error) {
	out := new(EmptyMessage)
	err := grpc.Invoke(ctx, "/paranoid.ParanoidNetwork/Truncate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paranoidNetworkClient) Utimes(ctx context.Context, in *UtimesRequest, opts ...grpc.CallOption) (*EmptyMessage, error) {
	out := new(EmptyMessage)
	err := grpc.Invoke(ctx, "/paranoid.ParanoidNetwork/Utimes", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paranoidNetworkClient) Chmod(ctx context.Context, in *ChmodRequest, opts ...grpc.CallOption) (*EmptyMessage, error) {
	out := new(EmptyMessage)
	err := grpc.Invoke(ctx, "/paranoid.ParanoidNetwork/Chmod", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ParanoidNetwork service

type ParanoidNetworkServer interface {
	// Used for health checking and discovery. Sends the IP and port of the
	// PFSD instance running on the client.
	Ping(context.Context, *PingRequest) (*EmptyMessage, error)
	// Filesystem calls
	Creat(context.Context, *CreatRequest) (*EmptyMessage, error)
	Write(context.Context, *WriteRequest) (*WriteResponse, error)
	Link(context.Context, *LinkRequest) (*EmptyMessage, error)
	Unlink(context.Context, *UnlinkRequest) (*EmptyMessage, error)
	Rename(context.Context, *RenameRequest) (*EmptyMessage, error)
	Truncate(context.Context, *TruncateRequest) (*EmptyMessage, error)
	Utimes(context.Context, *UtimesRequest) (*EmptyMessage, error)
	Chmod(context.Context, *ChmodRequest) (*EmptyMessage, error)
}

func RegisterParanoidNetworkServer(s *grpc.Server, srv ParanoidNetworkServer) {
	s.RegisterService(&_ParanoidNetwork_serviceDesc, srv)
}

func _ParanoidNetwork_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(ParanoidNetworkServer).Ping(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _ParanoidNetwork_Creat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(CreatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(ParanoidNetworkServer).Creat(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _ParanoidNetwork_Write_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(WriteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(ParanoidNetworkServer).Write(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _ParanoidNetwork_Link_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(LinkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(ParanoidNetworkServer).Link(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _ParanoidNetwork_Unlink_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(UnlinkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(ParanoidNetworkServer).Unlink(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _ParanoidNetwork_Rename_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(RenameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(ParanoidNetworkServer).Rename(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _ParanoidNetwork_Truncate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(TruncateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(ParanoidNetworkServer).Truncate(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _ParanoidNetwork_Utimes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(UtimesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(ParanoidNetworkServer).Utimes(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _ParanoidNetwork_Chmod_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(ChmodRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(ParanoidNetworkServer).Chmod(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _ParanoidNetwork_serviceDesc = grpc.ServiceDesc{
	ServiceName: "paranoid.ParanoidNetwork",
	HandlerType: (*ParanoidNetworkServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _ParanoidNetwork_Ping_Handler,
		},
		{
			MethodName: "Creat",
			Handler:    _ParanoidNetwork_Creat_Handler,
		},
		{
			MethodName: "Write",
			Handler:    _ParanoidNetwork_Write_Handler,
		},
		{
			MethodName: "Link",
			Handler:    _ParanoidNetwork_Link_Handler,
		},
		{
			MethodName: "Unlink",
			Handler:    _ParanoidNetwork_Unlink_Handler,
		},
		{
			MethodName: "Rename",
			Handler:    _ParanoidNetwork_Rename_Handler,
		},
		{
			MethodName: "Truncate",
			Handler:    _ParanoidNetwork_Truncate_Handler,
		},
		{
			MethodName: "Utimes",
			Handler:    _ParanoidNetwork_Utimes_Handler,
		},
		{
			MethodName: "Chmod",
			Handler:    _ParanoidNetwork_Chmod_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}

var fileDescriptor0 = []byte{
	// 487 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x94, 0x94, 0xcd, 0x6e, 0x13, 0x31,
	0x10, 0xc7, 0x49, 0xd8, 0x44, 0x61, 0xb2, 0x9b, 0x22, 0x57, 0x2d, 0xa1, 0x08, 0x84, 0xf6, 0x80,
	0x10, 0x87, 0x56, 0xd0, 0x03, 0xe2, 0x43, 0xe2, 0x80, 0xb8, 0x20, 0x40, 0x55, 0x05, 0xe2, 0x18,
	0xb9, 0xbb, 0x93, 0xd4, 0x6a, 0x6c, 0x2f, 0xb6, 0xab, 0x28, 0x4f, 0xc0, 0xd3, 0xf0, 0x8e, 0xf8,
	0x63, 0xb7, 0xf1, 0x42, 0xb3, 0x55, 0x8f, 0x33, 0x9e, 0xff, 0xcc, 0xdf, 0x9e, 0x9f, 0x0c, 0x4f,
	0x2a, 0xaa, 0xa8, 0x90, 0xac, 0x14, 0x68, 0x56, 0x52, 0x5d, 0x1c, 0x35, 0xf1, 0x61, 0xa5, 0xa4,
	0x91, 0x64, 0xd4, 0xc4, 0xf9, 0x04, 0xd2, 0x4f, 0xbc, 0x32, 0xeb, 0xaf, 0xa8, 0x35, 0x5d, 0x60,
	0xfe, 0x1e, 0xc6, 0x27, 0x4c, 0x2c, 0x4e, 0xf1, 0xd7, 0x25, 0x6a, 0x43, 0x00, 0xfa, 0xac, 0x9a,
	0xf6, 0x9e, 0xf6, 0x9e, 0xdf, 0x23, 0x29, 0x24, 0x95, 0x54, 0x66, 0xda, 0xf7, 0xd1, 0x2e, 0x8c,
	0x0b, 0xc9, 0xb9, 0x14, 0x33, 0x41, 0x39, 0x4e, 0xef, 0xba, 0x64, 0xfe, 0x12, 0xd2, 0x8f, 0x0a,
	0xa9, 0x69, 0xe4, 0x4e, 0x42, 0xcd, 0x79, 0xdd, 0xc0, 0x4a, 0x2a, 0x54, 0x9c, 0x69, 0xcd, 0xa4,
	0xd0, 0xbe, 0x4f, 0x96, 0x7f, 0x86, 0xf4, 0xa7, 0x62, 0x06, 0xaf, 0x97, 0xd8, 0xa8, 0xa4, 0x86,
	0xfa, 0xda, 0x94, 0x4c, 0x60, 0x28, 0xe7, 0x73, 0x8d, 0xc6, 0x8f, 0x4b, 0x5c, 0xbc, 0x44, 0xb1,
	0xb0, 0xd5, 0x89, 0x8b, 0xf3, 0x67, 0x90, 0xd5, 0xbd, 0x74, 0x65, 0x27, 0x20, 0xd9, 0x83, 0xec,
	0x6c, 0x6d, 0x50, 0xcf, 0x56, 0x36, 0x6d, 0x50, 0xf8, 0xae, 0x89, 0xb5, 0x39, 0xfe, 0xc2, 0xc4,
	0x45, 0x33, 0xf2, 0x3e, 0x8c, 0xe4, 0xb2, 0x9c, 0x45, 0x63, 0x6d, 0x46, 0xe0, 0x2a, 0x64, 0xfc,
	0x75, 0xf3, 0xc7, 0x90, 0xfd, 0x10, 0xcb, 0x48, 0xd4, 0xf2, 0x99, 0x1f, 0x43, 0x76, 0x8a, 0xee,
	0x21, 0x6e, 0xd3, 0xf3, 0x08, 0x76, 0xbe, 0xab, 0x4b, 0x51, 0xd0, 0x6d, 0xb7, 0xdf, 0xdc, 0xaf,
	0xef, 0x7d, 0xff, 0xee, 0x59, 0x17, 0x86, 0x71, 0xd4, 0xd7, 0xd7, 0xef, 0xc3, 0x84, 0x16, 0x85,
	0xdd, 0xe4, 0x4c, 0x63, 0x21, 0x45, 0x19, 0xde, 0x38, 0x21, 0x8f, 0x60, 0xb7, 0xce, 0x73, 0x56,
	0x28, 0xd9, 0x1c, 0x86, 0x47, 0xb4, 0x22, 0x2e, 0x4b, 0x36, 0x5f, 0x5f, 0x89, 0x92, 0x46, 0x54,
	0xe7, 0x5b, 0xa2, 0x81, 0x77, 0xf2, 0xc2, 0x2e, 0xfa, 0xdc, 0x1e, 0x6f, 0xdd, 0x9a, 0x3d, 0xc3,
	0xb0, 0xe1, 0x57, 0x7f, 0x12, 0xd8, 0x39, 0xa9, 0x79, 0xfb, 0x16, 0x78, 0x24, 0xaf, 0x21, 0x71,
	0x98, 0x91, 0xbd, 0xc3, 0x2b, 0x32, 0x23, 0xec, 0x0e, 0xf6, 0x37, 0xe9, 0x16, 0x9d, 0x77, 0xc8,
	0x1b, 0x18, 0x78, 0xc2, 0x48, 0x54, 0x12, 0x23, 0xd7, 0x21, 0x7d, 0x0b, 0x03, 0x4f, 0x47, 0x2c,
	0x8d, 0xd1, 0x3b, 0x78, 0xf0, 0x5f, 0x3e, 0x60, 0x64, 0xb5, 0xd6, 0xaf, 0x23, 0x26, 0xf6, 0x1b,
	0x11, 0xd4, 0x31, 0xf4, 0x1d, 0x0c, 0x03, 0x37, 0x24, 0xea, 0xde, 0x22, 0xa9, 0x5b, 0x1c, 0xa8,
	0x8a, 0xc5, 0x2d, 0xce, 0x3a, 0xc4, 0x1f, 0x60, 0xd4, 0xd0, 0x45, 0x1e, 0x6e, 0xaa, 0xfe, 0x21,
	0xee, 0x06, 0xeb, 0x1e, 0xb6, 0x96, 0xf5, 0x18, 0xbf, 0x1b, 0xf6, 0xe4, 0x00, 0x69, 0xed, 0x29,
	0x22, 0x66, 0xbb, 0xf4, 0x6c, 0xe8, 0xff, 0xa8, 0xe3, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0xbc,
	0x7b, 0xa5, 0xfb, 0xc5, 0x04, 0x00, 0x00,
}

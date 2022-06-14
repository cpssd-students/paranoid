// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: paranoid/discoverynetwork/v1/discovery.proto

package discoverynetwork

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DiscoveryNetworkServiceClient is the client API for DiscoveryNetworkService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DiscoveryNetworkServiceClient interface {
	// Discovery Calls
	Join(ctx context.Context, in *JoinRequest, opts ...grpc.CallOption) (*JoinResponse, error)
	Disconnect(ctx context.Context, in *DisconnectRequest, opts ...grpc.CallOption) (*DisconnectResponse, error)
}

type discoveryNetworkServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDiscoveryNetworkServiceClient(cc grpc.ClientConnInterface) DiscoveryNetworkServiceClient {
	return &discoveryNetworkServiceClient{cc}
}

func (c *discoveryNetworkServiceClient) Join(ctx context.Context, in *JoinRequest, opts ...grpc.CallOption) (*JoinResponse, error) {
	out := new(JoinResponse)
	err := c.cc.Invoke(ctx, "/paranoid.discoverynetwork.v1.DiscoveryNetworkService/Join", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *discoveryNetworkServiceClient) Disconnect(ctx context.Context, in *DisconnectRequest, opts ...grpc.CallOption) (*DisconnectResponse, error) {
	out := new(DisconnectResponse)
	err := c.cc.Invoke(ctx, "/paranoid.discoverynetwork.v1.DiscoveryNetworkService/Disconnect", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DiscoveryNetworkServiceServer is the server API for DiscoveryNetworkService service.
// All implementations must embed UnimplementedDiscoveryNetworkServiceServer
// for forward compatibility
type DiscoveryNetworkServiceServer interface {
	// Discovery Calls
	Join(context.Context, *JoinRequest) (*JoinResponse, error)
	Disconnect(context.Context, *DisconnectRequest) (*DisconnectResponse, error)
	mustEmbedUnimplementedDiscoveryNetworkServiceServer()
}

// UnimplementedDiscoveryNetworkServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDiscoveryNetworkServiceServer struct {
}

func (UnimplementedDiscoveryNetworkServiceServer) Join(context.Context, *JoinRequest) (*JoinResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Join not implemented")
}
func (UnimplementedDiscoveryNetworkServiceServer) Disconnect(context.Context, *DisconnectRequest) (*DisconnectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Disconnect not implemented")
}
func (UnimplementedDiscoveryNetworkServiceServer) mustEmbedUnimplementedDiscoveryNetworkServiceServer() {
}

// UnsafeDiscoveryNetworkServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DiscoveryNetworkServiceServer will
// result in compilation errors.
type UnsafeDiscoveryNetworkServiceServer interface {
	mustEmbedUnimplementedDiscoveryNetworkServiceServer()
}

func RegisterDiscoveryNetworkServiceServer(s grpc.ServiceRegistrar, srv DiscoveryNetworkServiceServer) {
	s.RegisterService(&DiscoveryNetworkService_ServiceDesc, srv)
}

func _DiscoveryNetworkService_Join_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiscoveryNetworkServiceServer).Join(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paranoid.discoverynetwork.v1.DiscoveryNetworkService/Join",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiscoveryNetworkServiceServer).Join(ctx, req.(*JoinRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiscoveryNetworkService_Disconnect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DisconnectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiscoveryNetworkServiceServer).Disconnect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paranoid.discoverynetwork.v1.DiscoveryNetworkService/Disconnect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiscoveryNetworkServiceServer).Disconnect(ctx, req.(*DisconnectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DiscoveryNetworkService_ServiceDesc is the grpc.ServiceDesc for DiscoveryNetworkService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DiscoveryNetworkService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "paranoid.discoverynetwork.v1.DiscoveryNetworkService",
	HandlerType: (*DiscoveryNetworkServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Join",
			Handler:    _DiscoveryNetworkService_Join_Handler,
		},
		{
			MethodName: "Disconnect",
			Handler:    _DiscoveryNetworkService_Disconnect_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "paranoid/discoverynetwork/v1/discovery.proto",
}
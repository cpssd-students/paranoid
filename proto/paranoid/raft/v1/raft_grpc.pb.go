// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: paranoid/raft/v1/raft.proto

package raft

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

// RaftNetworkServiceClient is the client API for RaftNetworkService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RaftNetworkServiceClient interface {
	AppendEntries(ctx context.Context, in *AppendEntriesRequest, opts ...grpc.CallOption) (*AppendEntriesResponse, error)
	RequestVote(ctx context.Context, in *RequestVoteRequest, opts ...grpc.CallOption) (*RequestVoteResponse, error)
	ClientToLeader(ctx context.Context, in *ClientToLeaderRequest, opts ...grpc.CallOption) (*ClientToLeaderResponse, error)
	Snapshot(ctx context.Context, in *SnapshotRequest, opts ...grpc.CallOption) (*SnapshotResponse, error)
}

type raftNetworkServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRaftNetworkServiceClient(cc grpc.ClientConnInterface) RaftNetworkServiceClient {
	return &raftNetworkServiceClient{cc}
}

func (c *raftNetworkServiceClient) AppendEntries(ctx context.Context, in *AppendEntriesRequest, opts ...grpc.CallOption) (*AppendEntriesResponse, error) {
	out := new(AppendEntriesResponse)
	err := c.cc.Invoke(ctx, "/paranoid.raft.v1.RaftNetworkService/AppendEntries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftNetworkServiceClient) RequestVote(ctx context.Context, in *RequestVoteRequest, opts ...grpc.CallOption) (*RequestVoteResponse, error) {
	out := new(RequestVoteResponse)
	err := c.cc.Invoke(ctx, "/paranoid.raft.v1.RaftNetworkService/RequestVote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftNetworkServiceClient) ClientToLeader(ctx context.Context, in *ClientToLeaderRequest, opts ...grpc.CallOption) (*ClientToLeaderResponse, error) {
	out := new(ClientToLeaderResponse)
	err := c.cc.Invoke(ctx, "/paranoid.raft.v1.RaftNetworkService/ClientToLeader", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftNetworkServiceClient) Snapshot(ctx context.Context, in *SnapshotRequest, opts ...grpc.CallOption) (*SnapshotResponse, error) {
	out := new(SnapshotResponse)
	err := c.cc.Invoke(ctx, "/paranoid.raft.v1.RaftNetworkService/Snapshot", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RaftNetworkServiceServer is the server API for RaftNetworkService service.
// All implementations must embed UnimplementedRaftNetworkServiceServer
// for forward compatibility
type RaftNetworkServiceServer interface {
	AppendEntries(context.Context, *AppendEntriesRequest) (*AppendEntriesResponse, error)
	RequestVote(context.Context, *RequestVoteRequest) (*RequestVoteResponse, error)
	ClientToLeader(context.Context, *ClientToLeaderRequest) (*ClientToLeaderResponse, error)
	Snapshot(context.Context, *SnapshotRequest) (*SnapshotResponse, error)
	mustEmbedUnimplementedRaftNetworkServiceServer()
}

// UnimplementedRaftNetworkServiceServer must be embedded to have forward compatible implementations.
type UnimplementedRaftNetworkServiceServer struct {
}

func (UnimplementedRaftNetworkServiceServer) AppendEntries(context.Context, *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendEntries not implemented")
}
func (UnimplementedRaftNetworkServiceServer) RequestVote(context.Context, *RequestVoteRequest) (*RequestVoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestVote not implemented")
}
func (UnimplementedRaftNetworkServiceServer) ClientToLeader(context.Context, *ClientToLeaderRequest) (*ClientToLeaderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClientToLeader not implemented")
}
func (UnimplementedRaftNetworkServiceServer) Snapshot(context.Context, *SnapshotRequest) (*SnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Snapshot not implemented")
}
func (UnimplementedRaftNetworkServiceServer) mustEmbedUnimplementedRaftNetworkServiceServer() {}

// UnsafeRaftNetworkServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RaftNetworkServiceServer will
// result in compilation errors.
type UnsafeRaftNetworkServiceServer interface {
	mustEmbedUnimplementedRaftNetworkServiceServer()
}

func RegisterRaftNetworkServiceServer(s grpc.ServiceRegistrar, srv RaftNetworkServiceServer) {
	s.RegisterService(&RaftNetworkService_ServiceDesc, srv)
}

func _RaftNetworkService_AppendEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppendEntriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftNetworkServiceServer).AppendEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paranoid.raft.v1.RaftNetworkService/AppendEntries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftNetworkServiceServer).AppendEntries(ctx, req.(*AppendEntriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftNetworkService_RequestVote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestVoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftNetworkServiceServer).RequestVote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paranoid.raft.v1.RaftNetworkService/RequestVote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftNetworkServiceServer).RequestVote(ctx, req.(*RequestVoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftNetworkService_ClientToLeader_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientToLeaderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftNetworkServiceServer).ClientToLeader(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paranoid.raft.v1.RaftNetworkService/ClientToLeader",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftNetworkServiceServer).ClientToLeader(ctx, req.(*ClientToLeaderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftNetworkService_Snapshot_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SnapshotRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftNetworkServiceServer).Snapshot(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paranoid.raft.v1.RaftNetworkService/Snapshot",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftNetworkServiceServer).Snapshot(ctx, req.(*SnapshotRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RaftNetworkService_ServiceDesc is the grpc.ServiceDesc for RaftNetworkService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RaftNetworkService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "paranoid.raft.v1.RaftNetworkService",
	HandlerType: (*RaftNetworkServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AppendEntries",
			Handler:    _RaftNetworkService_AppendEntries_Handler,
		},
		{
			MethodName: "RequestVote",
			Handler:    _RaftNetworkService_RequestVote_Handler,
		},
		{
			MethodName: "ClientToLeader",
			Handler:    _RaftNetworkService_ClientToLeader_Handler,
		},
		{
			MethodName: "Snapshot",
			Handler:    _RaftNetworkService_Snapshot_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "paranoid/raft/v1/raft.proto",
}

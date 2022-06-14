// Package pnetserver implements the ParanoidNetwork gRPC server.
// globals.go contains data used by each gRPC handler in pnetserver.
package pnetserver

import (
	pb "github.com/cpssd-students/paranoid/proto/paranoid/paranoidnetwork/v1"
)

var _ pb.ParanoidNetworkServiceServer = (*ParanoidServer)(nil)

// ParanoidServer implements the paranoidnetwork gRPC server
type ParanoidServer struct {
	pb.UnimplementedParanoidNetworkServiceServer
}

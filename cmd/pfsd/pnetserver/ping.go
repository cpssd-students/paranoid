package pnetserver

import (
	"context"
	"log"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/paranoidnetwork/v1"
)

// Ping implements the Ping RPC
func (s *ParanoidServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	node := globals.Node{
		IP:         req.Node.Ip,
		Port:       req.Node.Port,
		CommonName: req.Node.CommonName,
		UUID:       req.Node.Uuid,
	}
	log.Printf("Got Ping from %s", node)
	globals.Nodes.Add(node)
	globals.RaftNetworkServer.ChangeNodeLocation(req.Node.Uuid, req.Node.Ip, req.Node.Port)
	return &pb.PingResponse{}, nil
}

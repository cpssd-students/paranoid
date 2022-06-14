package pnetserver

import (
	"context"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	pb "github.com/cpssd-students/paranoid/proto/paranoidnetwork"
)

// Ping implements the Ping RPC
func (s *ParanoidServer) Ping(ctx context.Context, req *pb.Node) (*pb.EmptyMessage, error) {
	node := globals.Node{
		IP:         req.Ip,
		Port:       req.Port,
		CommonName: req.CommonName,
		UUID:       req.Uuid,
	}
	Log.Infof("Got Ping from %v", node)
	globals.Nodes.Add(node)
	globals.RaftNetworkServer.ChangeNodeLocation(req.Uuid, req.Ip, req.Port)
	return &pb.EmptyMessage{}, nil
}

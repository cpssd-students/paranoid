package dnetserver

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/cpssd-students/paranoid/proto/paranoid/discoverynetwork/v1"
)

// Disconnect method for Discovery Server
func (s *DiscoveryServer) Disconnect(
	ctx context.Context, req *pb.DisconnectRequest,
) (*pb.DisconnectResponse, error) {
	PoolLock.RLock()
	defer PoolLock.RUnlock()

	if _, ok := Pools[req.Pool]; ok {
		Pools[req.Pool].PoolLock.Lock()
		defer Pools[req.Pool].PoolLock.Unlock()
		err := checkPoolPassword(req.Pool, req.Password, req.Node)
		if err != nil {
			return &pb.DisconnectResponse{}, err
		}
	} else {
		log.Printf("Disconnect: Node %s (%s:%s) pool %s was not found",
			req.Node.Uuid, req.Node.Ip, req.Node.Port, req.Pool)
		returnError := status.Errorf(codes.NotFound, "pool %s was not found", req.Pool)
		return &pb.DisconnectResponse{}, returnError
	}

	if _, ok := Pools[req.Pool].Info.Nodes[req.Node.Uuid]; ok {
		delete(Pools[req.Pool].Info.Nodes, req.Node.Uuid)
		saveState(req.Pool)
		log.Printf("Disconnect: Node %s (%s:%s) disconnected",
			req.Node.Uuid, req.Node.Ip, req.Node.Port)
		return &pb.DisconnectResponse{}, nil
	}

	log.Printf("Disconnect: Node %s (%s:%s) was not found",
		req.Node.Uuid, req.Node.Ip, req.Node.Port)
	returnError := status.Errorf(codes.NotFound, "node was not found")
	return &pb.DisconnectResponse{}, returnError
}

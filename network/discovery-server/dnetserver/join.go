package dnetserver

import (
	pb "github.com/cpssd/paranoid/proto/discoverynetwork"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"log"
	"time"
)

// Join method for Discovery Server
func (s *DiscoveryServer) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	nodes := getNodes(req.Pool)
	response := pb.JoinResponse{RenewInterval.Nanoseconds() * 1000 * 1000, nodes}

	// Go through each node and check was the node there
	for _, node := range Nodes {
		if &node.Data == req.Node {
			if node.Active {
				log.Printf("[E] Join: node %s:%s is already part of the cluster\n", req.Node.Ip, req.Node.Port)
				returnError := grpc.Errorf(codes.AlreadyExists,
					"node is already part of the cluster")
				return &pb.JoinResponse{}, returnError
			}

			if node.Pool != req.Pool {
				log.Printf("[E] Join: node belongs to pool %s but tried to join pool %s\n", node.Pool, req.Pool)
				returnError := grpc.Errorf(codes.Internal,
					"node belongs to pool %s, but tried to join pool %s",
					node.Pool, req.Pool)
				return &pb.JoinResponse{}, returnError
			}

			node.Active = true
			return &response, nil
		}
	}

	newNode := Node{true, req.Pool, time.Now().Add(RenewInterval), *req.Node}
	Nodes = append(Nodes, newNode)

	return &response, nil
}

func getNodes(pool string) []*pb.Node {
	var nodes []*pb.Node
	for _, node := range Nodes {
		if node.Pool == pool {
			nodes = append(nodes, &node.Data)
		}
	}
	return nodes
}
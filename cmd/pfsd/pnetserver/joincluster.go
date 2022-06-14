package pnetserver

import (
	"context"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/pkg/raft"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/paranoidnetwork/v1"
)

// JoinCluster receives requests from nodes asking to join raft cluster
func (s *ParanoidServer) JoinCluster(
	ctx context.Context, req *pb.JoinClusterRequest,
) (*pb.JoinClusterResponse, error) {
	if req.PoolPassword == "" {
		if len(globals.PoolPasswordHash) != 0 {
			return &pb.JoinClusterResponse{},
				errors.New("cluster is password protected but no password was given")
		}
	} else {
		if err := bcrypt.CompareHashAndPassword(
			globals.PoolPasswordHash,
			append(globals.PoolPasswordSalt, []byte(req.PoolPassword)...),
		); err != nil {
			return &pb.JoinClusterResponse{},
				fmt.Errorf("unable to add node to raft cluster, password error: %v", err)
		}
	}

	node := raft.Node{
		IP:         req.Ip,
		Port:       req.Port,
		CommonName: req.CommonName,
		NodeID:     req.Uuid,
	}
	log.Printf("Got Ping from Node %s", node)
	err := globals.RaftNetworkServer.RequestAddNodeToConfiguration(node)
	if err != nil {
		return &pb.JoinClusterResponse{}, fmt.Errorf("unable to add node to raft cluster: %v", err)
	}
	return &pb.JoinClusterResponse{}, nil
}

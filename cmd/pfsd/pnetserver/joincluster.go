package pnetserver

import (
	"errors"
	"fmt"

	"paranoid/cmd/pfsd/globals"
	pb "paranoid/pkg/proto/paranoidnetwork"
	"paranoid/pkg/raft"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

// JoinCluster receives requests from nodes asking to join raft cluster
func (s *ParanoidServer) JoinCluster(ctx context.Context, req *pb.JoinClusterRequest) (*pb.EmptyMessage, error) {
	if req.PoolPassword == "" {
		if len(globals.PoolPasswordHash) != 0 {
			return &pb.EmptyMessage{}, errors.New("cluster is password protected but no password was given")
		}
	} else {
		err := bcrypt.CompareHashAndPassword(globals.PoolPasswordHash, append(globals.PoolPasswordSalt, []byte(req.PoolPassword)...))
		if err != nil {
			return &pb.EmptyMessage{}, fmt.Errorf("unable to add node to raft cluster, password error: %v", err)
		}
	}

	node := raft.Node{
		IP:         req.Ip,
		Port:       req.Port,
		CommonName: req.CommonName,
		NodeID:     req.Uuid,
	}
	Log.Infof("Got Ping from Node:", node)
	err := globals.RaftNetworkServer.RequestAddNodeToConfiguration(node)
	if err != nil {
		return &pb.EmptyMessage{}, fmt.Errorf("unable to add node to raft cluster: %v", err)
	}
	return &pb.EmptyMessage{}, nil
}

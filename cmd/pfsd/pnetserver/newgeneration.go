package pnetserver

import (
	"context"
	"log"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	pb "github.com/cpssd-students/paranoid/proto/paranoidnetwork"
)

// NewGeneration receives requests from nodes asking to create a new KeyPiece
// generation in preparation for joining the cluster.
func (s *ParanoidServer) NewGeneration(
	ctx context.Context, req *pb.NewGenerationRequest,
) (*pb.NewGenerationResponse, error) {
	if req.PoolPassword == "" {
		if len(globals.PoolPasswordHash) != 0 {
			return &pb.NewGenerationResponse{}, status.Errorf(codes.InvalidArgument,
				"cluster is password protected but no password was given")
		}
	} else {
		err := bcrypt.CompareHashAndPassword(globals.PoolPasswordHash,
			append(globals.PoolPasswordSalt, []byte(req.PoolPassword)...))
		if err != nil {
			return &pb.NewGenerationResponse{}, status.Errorf(codes.InvalidArgument,
				"unable to request new generation: password error: %v", err)
		}
	}

	log.Print("Requesting new generation")
	generationNumber, peers, err := globals.RaftNetworkServer.RequestNewGeneration(
		req.GetRequestingNode().Uuid)
	if err != nil {
		return &pb.NewGenerationResponse{},
			status.Errorf(codes.Unknown, "unable to create new generation")
	}
	return &pb.NewGenerationResponse{
		GenerationNumber: int64(generationNumber),
		Peers:            peers,
	}, nil
}

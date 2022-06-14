package pnetserver

import (
	"context"
	"errors"
	"log"
	"math/big"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/cmd/pfsd/keyman"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/paranoidnetwork/v1"
	raftpb "github.com/cpssd-students/paranoid/proto/paranoid/raft/v1"
)

// SendKeyPiece implements the SendKeyPiece RPC
func (s *ParanoidServer) SendKeyPiece(
	ctx context.Context, req *pb.KeyPieceSend,
) (*pb.SendKeyPieceResponse, error) {
	var prime big.Int
	prime.SetBytes(req.Key.Prime)
	// We must convert a slice to an array
	var fingerArray [32]byte
	copy(fingerArray[:], req.Key.ParentFingerprint)
	piece := &keyman.KeyPiece{
		Data:              req.Key.Data,
		ParentFingerprint: fingerArray,
		Prime:             &prime,
		Seq:               req.Key.Seq,
	}
	raftOwner := &raftpb.Node{
		Ip:         req.Key.OwnerNode.Ip,
		Port:       req.Key.OwnerNode.Port,
		CommonName: req.Key.OwnerNode.CommonName,
		NodeId:     req.Key.OwnerNode.Uuid,
	}
	raftHolder := &raftpb.Node{
		Ip:         globals.ThisNode.IP,
		Port:       globals.ThisNode.Port,
		CommonName: globals.ThisNode.CommonName,
		NodeId:     globals.ThisNode.UUID,
	}

	err := globals.HeldKeyPieces.AddPiece(req.Key.Generation, req.Key.OwnerNode.Uuid, piece)
	if err != nil {
		return &pb.SendKeyPieceResponse{},
			status.Errorf(codes.FailedPrecondition, "failed to save key piece to disk: %s", err)
	}
	log.Printf("Received KeyPiece from %s", req.Key.OwnerNode)
	if globals.RaftNetworkServer != nil &&
		globals.RaftNetworkServer.State.Configuration.HasConfiguration() &&
		req.AddElement {

		if err := globals.RaftNetworkServer.RequestKeyStateUpdate(
			raftOwner, raftHolder, req.Key.Generation,
		); err != nil {
			if errors.Is(err, keyman.ErrGenerationDeprecated) {
				return &pb.SendKeyPieceResponse{}, err
			}

			return &pb.SendKeyPieceResponse{},
				status.Errorf(codes.FailedPrecondition, "failed to commit to Raft: %s", err)
		}
	} else {
		return &pb.SendKeyPieceResponse{ClientMustCommit: true}, nil
	}
	return &pb.SendKeyPieceResponse{}, nil
}

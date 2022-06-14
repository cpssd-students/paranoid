package pnetserver

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/paranoidnetwork/v1"
)

// RequestKeyPiece implements the RequestKeyPiece RPC
func (s *ParanoidServer) RequestKeyPiece(
	ctx context.Context, req *pb.KeyPieceRequest,
) (*pb.KeyPiece, error) {
	key := globals.HeldKeyPieces.GetPiece(req.Generation, req.Node.Uuid)
	if key == nil {
		log.Printf("Key not found for node %s", req.Node)
		return &pb.KeyPiece{},
			status.Errorf(codes.NotFound, "Key not found for node %v", req.Node)
	}
	keyProto := &pb.KeyPiece{
		Data:              key.Data,
		ParentFingerprint: key.ParentFingerprint[:],
		Prime:             key.Prime.Bytes(),
		Seq:               key.Seq,
	}
	return keyProto, nil
}

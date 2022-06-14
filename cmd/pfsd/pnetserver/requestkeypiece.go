package pnetserver

import (
	"paranoid/cmd/pfsd/globals"
	pb "paranoid/proto/paranoidnetwork"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// RequestKeyPiece implements the RequestKeyPiece RPC
func (s *ParanoidServer) RequestKeyPiece(ctx context.Context, req *pb.KeyPieceRequest) (*pb.KeyPiece, error) {
	key := globals.HeldKeyPieces.GetPiece(req.Generation, req.Node.Uuid)
	if key == nil {
		Log.Warn("Key not found for node", req.Node)
		err := grpc.Errorf(codes.NotFound, "Key not found for node %v", req.Node)
		return &pb.KeyPiece{}, err
	}
	keyProto := &pb.KeyPiece{
		Data:              key.Data,
		ParentFingerprint: key.ParentFingerprint[:],
		Prime:             key.Prime.Bytes(),
		Seq:               key.Seq,
	}
	return keyProto, nil
}

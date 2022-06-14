package pnetclient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/cmd/pfsd/keyman"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/paranoidnetwork/v1"
)

// RequestKeyPiece from a node based on its UUID
func RequestKeyPiece(uuid string, generation int64) (*keyman.KeyPiece, error) {
	node, err := globals.Nodes.GetNode(uuid)
	if err != nil {
		return nil, errors.New("could not find node details")
	}

	conn, err := Dial(node)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %s", node, err)
	}
	defer conn.Close()

	client := pb.NewParanoidNetworkServiceClient(conn)

	thisNodeProto := &pb.Node{
		Ip:         globals.ThisNode.IP,
		Port:       globals.ThisNode.Port,
		CommonName: globals.ThisNode.CommonName,
		Uuid:       globals.ThisNode.UUID,
	}
	res, err := client.RequestKeyPiece(context.Background(), &pb.RequestKeyPieceRequest{
		Node:       thisNodeProto,
		Generation: generation,
	})
	if err != nil {
		log.Printf("Failed requesting KeyPiece from %s: %v", node, err)
		return nil, fmt.Errorf("failed requesting KeyPiece from %s: %w", node, err)
	}
	piece := res.Key

	log.Printf("Received KeyPiece from %s", node)
	var fingerprintArray [32]byte
	copy(fingerprintArray[:], piece.ParentFingerprint)
	var primeBig big.Int
	primeBig.SetBytes(piece.Prime)
	return &keyman.KeyPiece{
		Data:              piece.Data,
		ParentFingerprint: fingerprintArray,
		Prime:             &primeBig,
		Seq:               piece.Seq,
	}, nil
}

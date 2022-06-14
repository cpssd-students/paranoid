package pnetclient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/cmd/pfsd/keyman"
	pb "github.com/cpssd-students/paranoid/proto/paranoidnetwork"
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

	client := pb.NewParanoidNetworkClient(conn)

	thisNodeProto := &pb.Node{
		Ip:         globals.ThisNode.IP,
		Port:       globals.ThisNode.Port,
		CommonName: globals.ThisNode.CommonName,
		Uuid:       globals.ThisNode.UUID,
	}
	pieceProto, err := client.RequestKeyPiece(context.Background(), &pb.KeyPieceRequest{
		Node:       thisNodeProto,
		Generation: generation,
	},
	)
	if err != nil {
		log.Printf("Failed requesting KeyPiece from %s: %v", node, err)
		return nil, fmt.Errorf("failed requesting KeyPiece from %s: %w", node, err)
	}

	log.Printf("Received KeyPiece from %s", node)
	var fingerprintArray [32]byte
	copy(fingerprintArray[:], pieceProto.ParentFingerprint)
	var primeBig big.Int
	primeBig.SetBytes(pieceProto.Prime)
	piece := &keyman.KeyPiece{
		Data:              pieceProto.Data,
		ParentFingerprint: fingerprintArray,
		Prime:             &primeBig,
		Seq:               pieceProto.Seq,
	}
	return piece, nil
}

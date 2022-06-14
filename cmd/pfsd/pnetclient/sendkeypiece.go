package pnetclient

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/cmd/pfsd/keyman"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/paranoidnetwork/v1"
	raftpb "github.com/cpssd-students/paranoid/proto/paranoid/raft/v1"
)

// SendKeyPiece to the node specified by the UUID
func SendKeyPiece(uuid string, generation int64, piece *keyman.KeyPiece, addElement bool) error {
	node, err := globals.Nodes.GetNode(uuid)
	if err != nil {
		return errors.New("could not find node details")
	}

	conn, err := Dial(node)
	if err != nil {
		log.Printf("SendKeyPiece: failed to dial %s: %v", node, err)
		return fmt.Errorf("failed to dial: %s: %w", node, err)
	}
	defer conn.Close()

	client := pb.NewParanoidNetworkServiceClient(conn)

	thisNodeProto := &pb.Node{
		Ip:         globals.ThisNode.IP,
		Port:       globals.ThisNode.Port,
		CommonName: globals.ThisNode.CommonName,
		Uuid:       globals.ThisNode.UUID,
	}
	keyProto := &pb.KeyPiece{
		Data:              piece.Data,
		ParentFingerprint: piece.ParentFingerprint[:],
		Prime:             piece.Prime.Bytes(),
		Seq:               piece.Seq,
		Generation:        generation,
		OwnerNode:         thisNodeProto,
	}

	resp, err := client.SendKeyPiece(context.Background(), &pb.SendKeyPieceRequest{
		Key:        keyProto,
		AddElement: addElement,
	})
	if err != nil {
		log.Printf("Failed sending KeyPiece to %s: %v", node, err)
		return fmt.Errorf("Failed sending key piece to %s: %w", node, err)
	}

	if resp.ClientMustCommit && addElement {
		raftThisNodeProto := &raftpb.Node{
			Ip:         globals.ThisNode.IP,
			Port:       globals.ThisNode.Port,
			CommonName: globals.ThisNode.CommonName,
			NodeId:     globals.ThisNode.UUID,
		}
		raftOwnerNode := &raftpb.Node{
			Ip:         keyProto.GetOwnerNode().Ip,
			Port:       keyProto.GetOwnerNode().Port,
			CommonName: keyProto.GetOwnerNode().CommonName,
			NodeId:     keyProto.GetOwnerNode().Uuid,
		}
		if err := globals.RaftNetworkServer.RequestKeyStateUpdate(
			raftThisNodeProto,
			raftOwnerNode,
			generation,
		); err != nil {
			log.Printf("failed to commit to Raft: %v", err)
			return fmt.Errorf("failed to commit to Raft: %w", err)
		}
	}

	return nil
}

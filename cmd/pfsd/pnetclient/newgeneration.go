package pnetclient

import (
	"errors"
	"log"

	"context"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	pb "github.com/cpssd-students/paranoid/proto/paranoidnetwork"
)

// NewGeneration is used to create a new KeyPair generation in the cluster,
// prior to this node joining.
func NewGeneration(password string) (generation int64, peers []string, err error) {
	nodes := globals.Nodes.GetAll()
	for _, node := range nodes {
		log.Printf("Sending new generation request to node: %s", node)

		conn, err := Dial(node)
		if err != nil {
			log.Printf("NewGeneration: failed to dial %s: %v", node, err)
		}
		defer conn.Close()

		client := pb.NewParanoidNetworkClient(conn)

		resp, err := client.NewGeneration(context.Background(), &pb.NewGenerationRequest{
			RequestingNode: &pb.Node{
				Ip:         globals.ThisNode.IP,
				Port:       globals.ThisNode.Port,
				CommonName: globals.ThisNode.CommonName,
				Uuid:       globals.ThisNode.UUID,
			},
			PoolPassword: password,
		})
		if err != nil {
			log.Printf("Error requesting to create new generation %s: %v", node, err)
		} else {
			return resp.GenerationNumber, resp.Peers, nil
		}
	}
	return -1, nil, errors.New("unable to create new generation, no peer has returned okay")
}

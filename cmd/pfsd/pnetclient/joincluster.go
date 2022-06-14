package pnetclient

import (
	"errors"
	"log"

	"context"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/paranoidnetwork/v1"
)

//JoinCluster is used to request to join a raft cluster
func JoinCluster(password string) error {
	nodes := globals.Nodes.GetAll()
	for _, node := range nodes {
		log.Printf("Sending join cluster request to node %s", node)

		conn, err := Dial(node)
		if err != nil {
			log.Printf("JoinCluster: failed to dial %s: %v", node, err)
		}
		defer conn.Close()

		client := pb.NewParanoidNetworkClient(conn)

		_, err = client.JoinCluster(context.Background(), &pb.JoinClusterRequest{
			Ip:           globals.ThisNode.IP,
			Port:         globals.ThisNode.Port,
			CommonName:   globals.ThisNode.CommonName,
			Uuid:         globals.ThisNode.UUID,
			PoolPassword: password,
		})
		if err != nil {
			log.Printf("Error requesting to join cluster %s: %v", node, err)
		} else {
			return nil
		}
	}
	return errors.New("unable to join raft network, no peer has returned okay")
}

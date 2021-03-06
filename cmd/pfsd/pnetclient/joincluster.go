package pnetclient

import (
	"errors"

	"golang.org/x/net/context"

	"paranoid/cmd/pfsd/globals"

	pb "paranoid/pkg/proto/paranoidnetwork"
)

//JoinCluster is used to request to join a raft cluster
func JoinCluster(password string) error {
	nodes := globals.Nodes.GetAll()
	for _, node := range nodes {
		Log.Info("Sending join cluster request to node:", node)

		conn, err := Dial(node)
		if err != nil {
			Log.Error("JoinCluster: failed to dial ", node)
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
			Log.Error("Error requesting to join cluster", node, ":", err)
		} else {
			return nil
		}
	}
	return errors.New("unable to join raft network, no peer has returned okay")
}

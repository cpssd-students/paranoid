package dnetclient

import (
	"context"
	"errors"
	"log"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"

	pb "github.com/cpssd-students/paranoid/proto/paranoid/discoverynetwork/v1"
)

// Disconnect function used to disconnect from the server
func Disconnect(pool, password string) error {
	conn, err := dialDiscovery()
	if err != nil {
		log.Printf("Failed to dial discovery server at %s", globals.DiscoveryAddr)
		return errors.New("failed to dial discovery server")
	}
	defer conn.Close()

	dclient := pb.NewDiscoveryNetworkServiceClient(conn)

	_, err = dclient.Disconnect(context.Background(),
		&pb.DisconnectRequest{
			Pool:     pool,
			Password: password,
			Node: &pb.Node{
				Ip:   globals.ThisNode.IP,
				Port: globals.ThisNode.Port,
				Uuid: globals.ThisNode.UUID},
		})
	if err != nil {
		log.Print("Could not send disconnect message")
		return errors.New("could not send disconnect message")
	}

	return nil
}

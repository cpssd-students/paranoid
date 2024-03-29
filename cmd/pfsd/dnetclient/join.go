package dnetclient

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"

	pb "github.com/cpssd-students/paranoid/proto/paranoid/discoverynetwork/v1"
)

// Join function to call in order to join the server
// globals.ThisNode must be set before calling this.
func Join(pool, password string) error {
	conn, err := dialDiscovery()
	if err != nil {
		return fmt.Errorf("failed to dial discovery server: %s", err)
	}
	defer conn.Close()

	dclient := pb.NewDiscoveryNetworkServiceClient(conn)

	response, err := dclient.Join(context.Background(),
		&pb.JoinRequest{
			Pool:     pool,
			Password: password,
			Node: &pb.Node{
				Ip:         globals.ThisNode.IP,
				Port:       globals.ThisNode.Port,
				CommonName: globals.ThisNode.CommonName,
				Uuid:       globals.ThisNode.UUID,
			},
		})
	if err != nil {
		return fmt.Errorf("could not join the pool: %s", err)
	}

	interval := response.ResetInterval / 10 * 9
	globals.ResetInterval, err = time.ParseDuration(strconv.FormatInt(interval, 10) + "ms")
	if err != nil {
		log.Printf("Invalid renew interval: %v", err)
	}

	peerList := "Currently Connected: "
	for _, node := range response.Nodes {
		peerList += node.Ip + ":" + node.Port + ", "
		globals.Nodes.Add(globals.Node{
			IP:         node.Ip,
			Port:       node.Port,
			UUID:       node.Uuid,
			CommonName: node.CommonName,
		})
	}
	log.Println(peerList)

	log.Println("Successfully joined discovery network")

	return nil
}

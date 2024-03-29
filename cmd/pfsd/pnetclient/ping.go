package pnetclient

import (
	"context"
	"log"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/cmd/pfsd/upnp"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/paranoidnetwork/v1"
)

// Ping the peers
func Ping() {
	ip, err := upnp.GetIP()
	if err != nil {
		log.Printf("Can not ping peers: unable to get IP: %v", err)
	}

	nodes := globals.Nodes.GetAll()
	for _, node := range nodes {
		conn, err := Dial(node)
		if err != nil {
			log.Printf("Ping: failed to dial %s: %v", node, err)
		}
		defer conn.Close()

		client := pb.NewParanoidNetworkServiceClient(conn)

		if _, err := client.Ping(context.Background(), &pb.PingRequest{
			Node: &pb.Node{
				Ip:         ip,
				Port:       globals.ThisNode.Port,
				CommonName: globals.ThisNode.CommonName,
				Uuid:       globals.ThisNode.UUID,
			},
		}); err != nil {
			log.Printf("Can't ping %s: %v", node, err)
		}
	}
}

package pnetclient

import (
	"paranoid/cmd/pfsd/globals"
	"paranoid/cmd/pfsd/upnp"
	pb "paranoid/pkg/proto/paranoidnetwork"

	"golang.org/x/net/context"
)

// Ping the peers
func Ping() {
	ip, err := upnp.GetIP()
	if err != nil {
		Log.Fatal("Can not ping peers: unable to get IP. Error:", err)
	}

	nodes := globals.Nodes.GetAll()
	for _, node := range nodes {
		conn, err := Dial(node)
		if err != nil {
			Log.Error("Ping: failed to dial ", node)
		}
		defer conn.Close()

		client := pb.NewParanoidNetworkClient(conn)

		_, err = client.Ping(context.Background(), &pb.Node{
			Ip:         ip,
			Port:       globals.ThisNode.Port,
			CommonName: globals.ThisNode.CommonName,
			Uuid:       globals.ThisNode.UUID,
		})
		if err != nil {
			Log.Error("Can't ping", node)
		}
	}
}

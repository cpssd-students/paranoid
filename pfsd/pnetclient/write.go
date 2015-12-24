package pnetclient

import (
	"github.com/cpssd/paranoid/pfsd/globals"
	pb "github.com/cpssd/paranoid/proto/paranoidnetwork"
	"golang.org/x/net/context"
	"log"
)

func Write(path string, data []byte, offset, length uint64) {
	nodes := globals.Nodes.GetAll()
	for _, node := range nodes {
		conn, err := Dial(node)
		if err != nil {
			log.Println("Write error failed to dial ", node)
			continue
		}

		defer conn.Close()
		client := pb.NewParanoidNetworkClient(conn)

		_, err = client.Write(context.Background(), &pb.WriteRequest{path, data, offset, length})
		if err != nil {
			log.Println("Write Error on ", node, "Error:", err)
		}
	}
}

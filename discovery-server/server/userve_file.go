package server

import (
	"fmt"
	"github.com/pp2p/paranoid/discovery-server/dnetserver"
	pb "github.com/pp2p/paranoid/proto/fileserver"
	"golang.org/x/net/context"
)

func (s *FileserverServer) UnServeFile(ctx context.Context, req *pb.UnServeRequest) (*pb.ServeResponse, error) {
	for _, node := range dnetserver.Pools[req.Pool].Info.Nodes {
		if node.Uuid == req.Uuid {
			for key := range FileMap {
				if FileMap[key].FilePath == req.FilePath {
					fmt.Println(key, req.FilePath)
					delete(FileMap, key)
					return &pb.ServeResponse{"File Removed", ""}, nil
				}
			}
			return &pb.ServeResponse{"", "File Not Found"}, fmt.Errorf("Couldnt Find Key")
		}
	}
	return &pb.ServeResponse{"", "Node Not Found"}, fmt.Errorf("Couldnt Find Node")
}

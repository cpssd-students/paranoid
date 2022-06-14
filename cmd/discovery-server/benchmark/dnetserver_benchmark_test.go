package dnetservertest

import (
	"context"
	"log"
	"os"
	"path"
	"strconv"
	"testing"

	. "github.com/cpssd-students/paranoid/cmd/discovery-server/dnetserver"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/discoverynetwork/v1"
)

func TestMain(m *testing.M) {
	Pools = make(map[string]*Pool)
	StateDirectoryPath = path.Join(os.TempDir(), "server_state_bench")
	err := os.RemoveAll(StateDirectoryPath)
	if err != nil {
		log.Fatalf("Test setup failed: %v", err)
	}
	err = os.Mkdir(StateDirectoryPath, 0700)
	if err != nil {
		log.Fatalf("Test setup failed: %v", err)
	}
	os.Exit(m.Run())
}

func BenchmarkJoin(b *testing.B) {
	discovery := DiscoveryServer{}
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		joinRequest := &pb.JoinRequest{
			Node: &pb.Node{
				CommonName: "TestNode" + str,
				Ip:         "1.1.1." + str,
				Port:       "1001",
				Uuid:       "blahblah" + str,
			},
			Pool: "TestPool" + strconv.Itoa(n/5),
		}
		_, err := discovery.Join(context.Background(), joinRequest)
		if err != nil {
			log.Fatalf("Error joining network: %v", err)
		}
	}
}

func BenchmarkDisco(b *testing.B) {
	discovery := DiscoveryServer{}
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		joinRequest := &pb.JoinRequest{
			Node: &pb.Node{
				CommonName: "TestNode" + str,
				Ip:         "1.1.1.1" + str,
				Port:       "1001",
				Uuid:       "blahblah",
			},
			Pool: "TestPool" + strconv.Itoa(n/5),
		}
		_, _ = discovery.Join(context.Background(), joinRequest)
		disconnect := &pb.DisconnectRequest{
			Node: &pb.Node{
				CommonName: "TestNode" + str,
				Ip:         "1.1.1.1" + str,
				Port:       "1001",
				Uuid:       "blahblah",
			},
			Pool: "TestPool" + strconv.Itoa(n/5),
		}
		_, err := discovery.Disconnect(context.Background(), disconnect)
		if err != nil {
			log.Fatalf("Error disconnecting node 2: %v", err)
		}
	}
}
